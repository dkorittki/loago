package client

type WrappableError interface {
	error
	Unwrap() error
}

type CertificateReadError struct {
	Err error
}

func (e *CertificateReadError) Error() string {
	return "cannot read certificate"
}

func (e *CertificateReadError) Unwrap() error {
	return e.Err
}

type CertificateDecodeError struct{}

func (e *CertificateDecodeError) Error() string {
	return "cannot decode certificate"
}

type DialError struct {
	Err error
}

func (e *DialError) Error() string {
	return "cannot create connection"
}

func (e *DialError) Unwrap() error {
	return e.Err
}

type ActionError struct {
	Err error
}

func (e *ActionError) Error() string {
	return "error while performing action"
}

func (e *ActionError) Unwrap() error {
	return e.Err
}
