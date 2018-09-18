package modules

type SUpdateManager struct {
	ResourceManager
}

var (
	Updates SUpdateManager
)

func init() {

	Updates = SUpdateManager{NewAutoUpdateManager("update", "updates",
		// user view
		[]string{"localVersion", "remoteVersion", "status", "updateAvailable"},
		[]string{}, // admin view
	)}

	register(&Updates)
}
