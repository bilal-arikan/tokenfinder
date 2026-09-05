package ui

// Version is the semantic version shown in the tray tooltip and About box.
// Overridable at build time: -ldflags "-X tokenfinder/internal/ui.Version=1.2.3"
var Version = "1.0.1"

// RepoURL is shown in the About box.
const RepoURL = "https://github.com/bilal-arikan/tokenfinder"
