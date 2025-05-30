// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// UpdateInstancePoolDetails The data to update an instance pool.
type UpdateInstancePoolDetails struct {

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the instance configuration associated with the
	// instance pool.
	InstanceConfigurationId *string `mandatory:"false" json:"instanceConfigurationId"`

	// The placement configurations for the instance pool. Provide one placement configuration for
	// each availability domain.
	// To use the instance pool with a regional subnet, provide a placement configuration for
	// each availability domain, and include the regional subnet in each placement
	// configuration.
	PlacementConfigurations []UpdateInstancePoolPlacementConfigurationDetails `mandatory:"false" json:"placementConfigurations"`

	// The number of instances that should be in the instance pool.
	// To determine whether capacity is available for a specific shape before you resize an instance pool,
	// use the CreateComputeCapacityReport
	// operation.
	Size *int `mandatory:"false" json:"size"`

	// A user-friendly formatter for the instance pool's instances. Instance displaynames follow the format.
	// The formatter does not retroactively change instance's displaynames, only instance displaynames in the future follow the format
	InstanceDisplayNameFormatter *string `mandatory:"false" json:"instanceDisplayNameFormatter"`

	// A user-friendly formatter for the instance pool's instances. Instance hostnames follow the format.
	// The formatter does not retroactively change instance's hostnames, only instance hostnames in the future follow the format
	InstanceHostnameFormatter *string `mandatory:"false" json:"instanceHostnameFormatter"`
}

func (m UpdateInstancePoolDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateInstancePoolDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
