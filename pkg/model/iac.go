package model

// ValuePair represents a key value pair and is a workaround for viper converting all keys to lower case.
// https://github.com/spf13/viper/issues/371
type ValuePair struct {
	// Key is the case sensitive argument or environment name.
	Key string

	// Value is the associated value.
	Value string
}

// BinaryConfig represent the configuration in order to execute a terraform command.
type BinaryConfig struct {
	// Binary is the path to the binary to be Ran.
	// This will will search the PATH for which binary to use.
	// For more information refer to https://golang.org/pkg/os/exec/#LookPath
	// If empty `terraform` will be used.
	Binary string

	// Args are optional arguments to be supplied to the binary. Note this will show up in plain text.
	Args []ValuePair

	// PrivateArgs are arguments that will supplied to the binary via Environment Variables with the Prefix TF_VAR_.
	// Refer to https://www.terraform.io/docs/cli/config/environment-variables.html#tf_var_name for more information.
	PrivateArgs []ValuePair

	// Environment is additional environment variables to give the binary to run on top of the current environment.
	Environment []ValuePair

	// WorkingDirectory is the working directory to run the specified Binary.
	// If empty the current directory will be used.
	WorkingDirectory string
}
