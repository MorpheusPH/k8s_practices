package v1beta1

const (
	// Ready => controller considers this resource Ready
	Ready = "Ready"
	//Config
	Config = "Config"
	// Parsed
	Parsed = "Parsed"
	// Revision
	Revision = "Revision"
	// Render
	Workflow = "Workflow"
	// Task
	Task = "Task"
	// Render
	Render = "Render"
	//Validated
	Validated = "Validated"
	// Error => last recorded error
	Error = "Error"
)

const (
	InitReason                 string = "Init"
	SuccessdedReason           string = "Succeeded"
	FailedReason               string = "Failed"
	GetLastReleaseFailedReason string = "GetLastReleaseFailed"
	ProgressingReason          string = "Progressing"
)
