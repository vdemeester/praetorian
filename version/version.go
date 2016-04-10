package version

// Name is executable name of this application.
const Name = "praetorian"

// Description is the simple description of this application
const Description = "A ssh praetorian (bouncer, minder or whatever) ; it's just a cool restricted command script."

// Version is version string of this application.
// Version is changed with semantic versioning.
const Version string = "0.5.0-dev"

// GitCommit is the commit hash used to build the praetorian binary.
// It will be overriden automatically by the build system.
var GitCommit = "HEAD"
