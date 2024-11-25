package logger

const (
	LogDefaultFormat = "%s: %s"

	ErrLogFailedWrite = "Failed write to log"

	ErrMetricInvalidType         = "invalid metric type"
	ErrMetricNotFound            = "metric not found"
	ErrMetricInvalidGaugeValue   = "Invalid gauge value"
	ErrMetricInvalidCounterValue = "Invalid counter value"
	ErrWriteResponce             = "Failed to write response"
	OkUpdated                    = "updated successfully"

	ErrHTMLTemplateParse   = "Failed to parse template"
	ErrHTMLTemplateExecute = "Failed to execute template"

	ErrFlagsParse = "Failed to parse arguments"

	ErrServerInternalError = "Internal server error"
	ErrServerNotStarted    = "Server can't be started"
	OkServerStarted        = "Server started"

	ErrAgentResponseCode  = "Unexpected response code"
	ErrAgentCreateRequest = "Error creating request"
	ErrAgentSendRequest   = "Error sending request"
	ErrAgentCloseRequest  = "Error closing response body"

	ErrFlagUnknown      = "Unknown flags provided"
	ErrFlagInvalidValue = "Invalid flag value"
)
