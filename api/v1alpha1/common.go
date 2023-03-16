package v1alpha1

const (
	AnnotationsOperationId = "management.azure.com/operationId"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ProvisioningState describes the job completion status.
// Only one of the following provisioning status may be specified.
// +kubebuilder:validation:Enum=Provisioning;Succeeded;Failed;
type ProvisioningState string

const (
	// Provisioning indicates resource provisioning is underway.
	Provisioning ProvisioningState = "Provisioning"

	// Succeeded indicates resource is provisioned successfully.
	Succeeded ProvisioningState = "Succeeded"

	// Failed indicates resource provisioning failed.
	Failed ProvisioningState = "Failed"
)

const Unknown = "Unknown"

// ProvisioningStatus describes the provisioning status of the resource.
type ProvisioningStatus struct {
	Status      ProvisioningState  `json:"status,omitempty"`
	OperationID string             `json:"operationID,omitempty"`
	Error       *ProvisioningError `json:"error,omitempty"`
}
