package v20170831

import "encoding/json"

// The validate tag is used for validation
// Reference to gopkg.in/go-playground/validator.v9

// ResourcePurchasePlan defines resource plan as required by ARM
// for billing purposes.
type ResourcePurchasePlan struct {
	Name          string `json:"name,omitempty"`
	Product       string `json:"product,omitempty"`
	PromotionCode string `json:"promotionCode,omitempty"`
	Publisher     string `json:"publisher,omitempty"`
}

// HostedMaster complies with the ARM model of
// resource definition in a JSON template.
type HostedMaster struct {
	ID       string                `json:"id,omitempty"`
	Location string                `json:"location,omitempty" validate:"required"`
	Name     string                `json:"name,omitempty"`
	Plan     *ResourcePurchasePlan `json:"plan,omitempty"`
	Tags     map[string]string     `json:"tags,omitempty"`
	Type     string                `json:"type,omitempty"`

	Properties *Properties `json:"properties"`
}

// Properties represents the ACS cluster definition
type Properties struct {
	ProvisioningState ProvisioningState   `json:"provisioningState,omitempty"`
	Version           OrchestratorVersion `json:"version"`
	DNSPrefix         string              `json:"dnsPrefix" validate:"required"`
	// Master LB public endpoint/FQDN with port
	// The format will be FQDN:2376
	// Not used during PUT, returned as part of GET
	FQDN                    string                   `json:"fqdn,omitempty"`
	AgentPoolProfiles       []*AgentPoolProfile      `json:"agentPoolProfiles,omitempty" validate:"dive,required"`
	LinuxProfile            *LinuxProfile            `json:"linuxProfile,omitempty" validate:"required"`
	WindowsProfile          *WindowsProfile          `json:"windowsProfile,omitempty"`
	ServicePrincipalProfile *ServicePrincipalProfile `json:"servicePrincipalProfile,omitempty"`
	NetworkProfile          *NetworkProfile          `json:"networkProfile,omitempty"`
	AccessProfiles          []*AccessProfile         `json:"accessProfiles,omitempty"`
	JumpboxProfile          *JumpboxProfile          `json:"jumpboxProfile,omitempty"`
}

// ServicePrincipalProfile contains the client and secret used by the cluster for Azure Resource CRUD
// The 'Secret' parameter could be either a plain text, or referenced to a secret in a keyvault.
// In the latter case, the format of the parameter's value should be
// "/subscriptions/<SUB_ID>/resourceGroups/<RG_NAME>/providers/Microsoft.KeyVault/vaults/<KV_NAME>/secrets/<NAME>[/<VERSION>]"
// where:
//    <SUB_ID> is the subscription ID of the keyvault
//    <RG_NAME> is the resource group of the keyvault
//    <KV_NAME> is the name of the keyvault
//    <NAME> is the name of the secret.
//    <VERSION> (optional) is the version of the secret (default: the latest version)
type ServicePrincipalProfile struct {
	ClientID string `json:"clientId,omitempty" validate:"required"`
	Secret   string `json:"secret,omitempty" validate:"required"`
}

// LinuxProfile represents the Linux configuration passed to the cluster
type LinuxProfile struct {
	AdminUsername string `json:"adminUsername" validate:"required"`

	SSH struct {
		PublicKeys []PublicKey `json:"publicKeys" validate:"required,len=1"`
	} `json:"ssh" validate:"required"`
}

// PublicKey represents an SSH key for LinuxProfile
type PublicKey struct {
	KeyData string `json:"keyData"`
}

// WindowsProfile represents the Windows configuration passed to the cluster
type WindowsProfile struct {
	AdminUsername string `json:"adminUsername,omitempty" validate:"required"`
	AdminPassword string `json:"adminPassword,omitempty" validate:"required"`
}

// ProvisioningState represents the current state of container service resource.
type ProvisioningState string

const (
	// Creating means ContainerService resource is being created.
	Creating ProvisioningState = "Creating"
	// Updating means an existing ContainerService resource is being updated
	Updating ProvisioningState = "Updating"
	// Failed means resource is in failed state
	Failed ProvisioningState = "Failed"
	// Succeeded means resource created succeeded during last create/update
	Succeeded ProvisioningState = "Succeeded"
	// Deleting means resource is in the process of being deleted
	Deleting ProvisioningState = "Deleting"
	// Migrating means resource is being migrated from one subscription or
	// resource group to another
	Migrating ProvisioningState = "Migrating"
)

// NetworkProfile represents the definition of network profile for the cluster
type NetworkProfile struct {
	// PodCidr defines the range of addresses that will be used for Pods.
	// Allocation of subnets to agents will be handed by kube-controller-manager
	PodCidr string `json:"podCidr,omitempty"`
	// ServiceCidr defines the range for Kubernetes services
	ServiceCidr string `json:"serviceCidr,omitempty"`
	// AgentCidr defines the range for Agent VMs in the subnet
	AgentCidr        string `json:"agentCidr,omitempty"`
	KubeDNSServiceIP string `json:"kubeDNSServiceIP,omitempty"`
}

// JumpboxProfile represents the definition of bastion machine to access the agent nodes
type JumpboxProfile struct {
	PublicIPAddressID string `json:"publicIPAddressID,omitempty"`
	PublicIPAddress   string `json:"publicIPAddress,omitempty"`
	InternalIPAddress string `json:"internalIPAddress,omitempty"`
}

// AgentPoolProfile represents configuration of VMs running agent
// daemons that register with the master and offer resources to
// host applications in containers.
type AgentPoolProfile struct {
	Name           string `json:"name" validate:"required"`
	Count          int    `json:"count" validate:"required,min=1,max=100"`
	VMSize         string `json:"vmSize" validate:"required"`
	OSDiskSizeGB   int    `json:"osDiskSizeGB,omitempty" validate:"min=0,max=1023"`
	StorageProfile string `json:"storageProfile" validate:"eq=StorageAccount|eq=ManagedDisks|len=0"`
	VnetSubnetID   string `json:"vnetSubnetID,omitempty"`

	// OSType is the operating system type for agents
	// Set as nullable to support backward compat because
	// this property was added later.
	// If the value is null or not set, it defaulted to Linux.
	OSType OSType `json:"osType,omitempty"`

	// subnet is internal
	subnet string
}

// AccessProfile represents role name and kubeconfig
type AccessProfile struct {
	RoleName   string `json:"roleName"`
	KubeConfig string `json:"kubeConfig"`
}

// UnmarshalJSON unmarshal json using the default behavior
// And do fields manipulation, such as populating default value
func (a *AgentPoolProfile) UnmarshalJSON(b []byte) error {
	// Need to have a alias type to avoid circular unmarshal
	type aliasAgentPoolProfile AgentPoolProfile
	aa := aliasAgentPoolProfile{}
	if e := json.Unmarshal(b, &aa); e != nil {
		return e
	}
	*a = AgentPoolProfile(aa)
	if a.Count == 0 {
		// if AgentPoolProfile.Count is missing or 0, set it to default 1
		a.Count = 1
	}

	if a.StorageProfile == "" {
		// if StorageProfile is missing, set to default StorageAccount
		a.StorageProfile = StorageAccount
	}

	if string(a.OSType) == "" {
		// OSType is the operating system type for agents
		// Set as nullable to support backward compat because
		// this property was added later.
		// If the value is null or not set, it defaulted to Linux.
		a.OSType = Linux
	}

	// OSDiskSizeGB is an override value. vm sizes have default OS disk sizes.
	// If it is not set. The user should get the default for the vm size
	return nil
}

// OrchestratorVersion defines the version for orchestratorType
type OrchestratorVersion string

// OSType represents OS types of agents
type OSType string

// HasWindows returns true if the cluster contains windows
func (a *Properties) HasWindows() bool {
	for _, agentPoolProfile := range a.AgentPoolProfiles {
		if agentPoolProfile.OSType == Windows {
			return true
		}
	}
	return false
}

// IsCustomVNET returns true if the customer brought their own VNET
func (a *AgentPoolProfile) IsCustomVNET() bool {
	return len(a.VnetSubnetID) > 0
}

// IsWindows returns true if the agent pool is windows
func (a *AgentPoolProfile) IsWindows() bool {
	return a.OSType == Windows
}

// IsLinux returns true if the agent pool is linux
func (a *AgentPoolProfile) IsLinux() bool {
	return a.OSType == Linux
}

// GetSubnet returns the read-only subnet for the agent pool
func (a *AgentPoolProfile) GetSubnet() string {
	return a.subnet
}

// SetSubnet sets the read-only subnet for the agent pool
func (a *AgentPoolProfile) SetSubnet(subnet string) {
	a.subnet = subnet
}

// IsManagedDisks returns true if the customer specified managed disks
func (a *AgentPoolProfile) IsManagedDisks() bool {
	return a.StorageProfile == ManagedDisks
}

// IsStorageAccount returns true if the customer specified storage account
func (a *AgentPoolProfile) IsStorageAccount() bool {
	return a.StorageProfile == StorageAccount
}
