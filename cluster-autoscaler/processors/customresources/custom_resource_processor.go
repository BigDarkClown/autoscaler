/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package customresources

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

const (
	// MetricsGenericResource - for when there is no information about resource type
	MetricsGenericResource = "generic"
	// MetricsMissingResource - for when there's a label, but resource didn't appear
	MetricsMissingResource = "missing-resource"
	// MetricsUnexpectedResourceLabel - for when there's a label, but no resource at all
	MetricsUnexpectedResourceLabel = "unexpected-label"
	// MetricsUnknownResource - for when resource type is unknown
	MetricsUnknownResource = "not-listed"
	// MetricsErrorResource - for when there was an error obtaining resource type
	MetricsErrorResource = "error"
	// MetricsNoResource - for when there is no resource and no label all
	MetricsNoResource = ""
)

// CustomResourceProcessor is interface defining handling custom resources
type CustomResourceProcessor interface {
	// FilterOutNodesWithUnreadyResource removes nodes that should have a resource, but don't have it in allocatable
	// from ready nodes list and updates their status to unready on all nodes list.
	FilterOutNodesWithUnreadyResource(context *context.AutoscalingContext, resourceName apiv1.ResourceName,
		resourceLabel string, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node)
	// GetResourceTypeForMetrics returns name of the resource used on the node if resource is present
	// empty string if there is no resource an "generic" if resource type is unknown.
	GetResourceTypeForMetrics(context *context.AutoscalingContext, resourceName apiv1.ResourceName,
		resourceLabel string, availableResourceTypes map[string]struct{}, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) string
	// GetNodeTargetResource returns the number of custom resources on a given node.
	GetNodeTargetResource(context *context.AutoscalingContext, resourceName apiv1.ResourceName,
		resourceLabel string, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) (resourceType string, resourceCount int64, error errors.AutoscalerError)
	// CleanUp cleans up processor's internal structures.
	CleanUp()
}

type DriverCustomResourceProcessor struct {
}

// FilterOutNodesWithUnreadyResource removes nodes that should have a resource, but don't have it in allocatable
// from ready nodes list and updates their status to unready on all nodes list.
// This is a hack/workaround for nodes with custom resources coming up without installed drivers, resulting
// in resource missing from their allocatable and capacity.
func (p *DriverCustomResourceProcessor) FilterOutNodesWithUnreadyResource(context *context.AutoscalingContext, resourceName apiv1.ResourceName,
	resourceLabel string, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithUnreadyResource := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		_, hasResourceLabel := node.Labels[resourceLabel]
		resourceAllocatable, hasResourceAllocatable := node.Status.Allocatable[resourceName]
		// We expect node to have resource based on label, but it doesn't show up on node
		// object. Assume the node is still not fully started (installing resource drivers).
		if hasResourceLabel && (!hasResourceAllocatable || resourceAllocatable.IsZero()) {
			klog.V(3).Infof("Overriding status of node %v, which seems to have unready %v resource",
				node.Name, resourceName)
			nodesWithUnreadyResource[node.Name] = kubernetes.GetUnreadyNodeCopy(node)
		} else {
			newReadyNodes = append(newReadyNodes, node)
		}
	}
	// Override any node with unready resource with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithUnreadyResource[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}

// GetResourceTypeForMetrics returns name of the resource used on the node if resource is present
// empty string if there is no resource an "generic" if resource type is unknown.
func (p *DriverCustomResourceProcessor) GetResourceTypeForMetrics(context *context.AutoscalingContext, resourceName apiv1.ResourceName,
	resourceLabel string, availableResourceTypes map[string]struct{}, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) string {
	// we use the resource label if there is one
	resourceType, labelFound := node.Labels[resourceLabel]
	capacity, capacityFound := node.Status.Capacity[resourceName]

	// no label, fallback to generic solution
	if !labelFound {
		if capacityFound && !capacity.IsZero() {
			return MetricsGenericResource
		}
		return MetricsNoResource
	}

	// resource label & capacity are present - consistent state
	if capacityFound {
		if _, found := availableResourceTypes[resourceType]; found {
			return resourceType
		}
		return MetricsUnknownResource
	}

	// resource label present but no capacity (yet?) - check the node template if possible
	if nodeGroup != nil {
		template, err := nodeGroup.TemplateNodeInfo()
		if err != nil {
			klog.Warningf("Failed to build template for getting %v resource metrics for node %v: %v", resourceName, node.Name, err)
			return MetricsErrorResource
		}

		if _, found := template.Node().Status.Capacity[resourceName]; found {
			return MetricsMissingResource
		}

		// if template does not define resource we assume node will not have any even if it has resource label
		klog.Warningf("Template does not define %v resource even though node from its node group does; node=%v", resourceName, node.Name)
		return MetricsUnexpectedResourceLabel
	}

	return MetricsUnexpectedResourceLabel
}

// GetNodeTargetResource returns the number of custom resources on a given node.
// This includes resources which are not yet ready to use and visible in kubernetes.
func (p *DriverCustomResourceProcessor) GetNodeTargetResource(context *context.AutoscalingContext, resourceName apiv1.ResourceName,
	resourceLabel string, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) (resourceType string, resourceCount int64, error errors.AutoscalerError) {
	resourceLabel, found := node.Labels[resourceLabel]
	if !found {
		return "", 0, nil
	}

	resourceAllocatable, found := node.Status.Allocatable[resourceName]
	if found && resourceAllocatable.Value() > 0 {
		return resourceLabel, resourceAllocatable.Value(), nil
	}

	// A node is supposed to have resource (based on label), but they're not available yet
	// (driver haven't installed yet?). Unfortunately we can't deduce how many resources it
	// will actually have from labels (just that it will have some).
	// Ready for some evil hacks? Well, you won't be disappointed - let's pretend we haven't
	// seen the node and use the template we use for scale from 0. It'll be our little secret.

	if nodeGroup == nil {
		// We expect this code path to be triggered by situation when we are looking at a node which is
		// expected to have resources (has resource label), but those are not yet visible in node's resource
		// (e.g. resource drivers are still being installed).
		// In case of node coming from autoscaled node group we would look and node group template here.
		// But for nodes coming from non-autoscaled groups we have no such possibility.
		// Let's hope it is a transient error. As long as it exists we will not scale nodes groups with resources.
		return "", 0, errors.NewAutoscalerError(errors.InternalError,
			fmt.Sprintf("node without with %v resource label, without capacity not belonging to autoscaled node group", resourceName))
	}

	template, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		klog.Errorf("Failed to build template for getting %v resource estimation for node %v: %v", resourceName, node.Name, err)
		return "", 0, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}
	if resourceCapacity, found := template.Node().Status.Capacity[resourceName]; found {
		return resourceLabel, resourceCapacity.Value(), nil
	}

	// if template does not define resources we assume node will not have any even if ith has resource label
	klog.Warningf("Template does not define %v resource even though node from its node group does; node=%v", resourceName, node.Name)
	return "", 0, nil
}

func (p *DriverCustomResourceProcessor) CleanUp() {
}

// NewDefaultCustomResourceProcessor returns a default instance of CustomResourceProcessor.
func NewDefaultCustomResourceProcessor() CustomResourceProcessor {
	return &DriverCustomResourceProcessor{}
}
