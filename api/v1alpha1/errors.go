package v1alpha1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ErrorCode string
type Severity string

type ErrorObj struct {
	Id                int         `json:"id"`
	Code              ErrorCode   `json:"code"`
	PossibleCause     string      `json:"possibleCause,omitempty"`
	RecommendedAction string      `json:"recommendedAction,omitempty"`
	Severity          Severity    `json:"severity"`
	Message           string      `json:"message,omitempty"`
	SummaryMessage    string      `json:"summaryMessage,omitempty"`
	RunAsAccountId    string      `json:"runAsAccountId,omitempty"`
	UpdatedTimeStamp  metav1.Time `json:"updatedTimeStamp,omitempty"`
}

type ProvisioningError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (eo ErrorObj) ProvisioningError() *ProvisioningError {
	return &ProvisioningError{Code: eo.Code, Message: fmt.Sprintf("Summary: %s, message: %s", eo.SummaryMessage, eo.Message)}
}

const (
	Information Severity = "Information"
	Warning     Severity = "Warning"
	Error       Severity = "Error"
)

const (
	ConnectionErrorCode  ErrorCode = "ConnectionError"
	SshCmdErrorCode      ErrorCode = "SshCmdError"
	PermissionErrorCode  ErrorCode = "PermissionError"
	CredentialErrorCode  ErrorCode = "CredentialError"
	GeneralErrorCode     ErrorCode = "GeneralError"
	RefreshSiteErrorCode ErrorCode = "RefreshSiteErrorCode"
	OverMaxTimeErrorCode ErrorCode = "OverMaxTimeErrorCode"
	RejectedErrorCode    ErrorCode = "RejectedErrorCode"
	MultipleErrorCode    ErrorCode = "MultipleErrorCode"
)

var (
	ConnectionError  = ErrorObj{Id: 1, Code: ConnectionErrorCode, SummaryMessage: "Cannot connect to target server", PossibleCause: "Server is down", RecommendedAction: "Check whether target server is up and running", Severity: Error}
	SshCmdError      = ErrorObj{Id: 2, Code: SshCmdErrorCode, SummaryMessage: "Shell command execution failure error", PossibleCause: "Server is running a non-compatible linux distribution", RecommendedAction: "Ensure target server is linux based server, or report this error to azure migrate team", Severity: Error}
	PermissionError  = ErrorObj{Id: 3, Code: PermissionErrorCode, SummaryMessage: "Shell command execution permission denied", PossibleCause: "Do not have enough permission to run command on target server", RecommendedAction: "Grant permission to the credentials", Severity: Error}
	CredentialError  = ErrorObj{Id: 4, Code: CredentialErrorCode, SummaryMessage: "Credential error", PossibleCause: "Didn't have properly set the credentials", RecommendedAction: "Check you credential settings", Severity: Error}
	GeneralError     = ErrorObj{Id: 5, Code: GeneralErrorCode, SummaryMessage: "Unknown error", PossibleCause: "", RecommendedAction: "Refer to the error message or report to azure migration team", Severity: Error}
	RefreshSiteError = ErrorObj{Id: 6, Code: RefreshSiteErrorCode, SummaryMessage: "Discovery failure", PossibleCause: "Error occurred while refresh site", RecommendedAction: "Report this to azure migrate team", Severity: Error}
	OverMaxTimeError = ErrorObj{Id: 7, Code: OverMaxTimeErrorCode, SummaryMessage: "Discovery over max time", PossibleCause: "Discovery running slow or too many servers to be discovered", RecommendedAction: "Re-trigger the action again, or reduce the number of servers to be discovered", Severity: Warning}
	RejectedError    = ErrorObj{Id: 8, Code: RejectedErrorCode, SummaryMessage: "Already had an active discovery job running", PossibleCause: "Rejected due to already had an active job is running", RecommendedAction: "Wait until the active job completed", Severity: Warning}
	MultipleError    = ErrorObj{Id: 9, Code: MultipleErrorCode, SummaryMessage: "Multiple error occurred", PossibleCause: "Multiple errors occurred during discovery", RecommendedAction: "Check error details on target server", Severity: Error}
)
