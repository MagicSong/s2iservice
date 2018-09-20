package httphandler

import "fmt"

var (
	ErrInvalidGitURL       = fmt.Errorf("Invalid git source")
	ErrBadAuthorization    = fmt.Errorf("Bad  Authorization")
	ErrInvalidCallbackURL  = fmt.Errorf("Illegal Callback URL")
	ErrInvaildImageName    = fmt.Errorf("Illegal Image Name")
	ErrLackOfRuntimeOption = fmt.Errorf("RuntimeImage and RuntimeArtifact must co-exsit")
	ErrResourceNotFound    = fmt.Errorf("Could not find the specified resource")
)
