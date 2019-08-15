package main

import "github.com/jeshan/tfbridge/tfbridge/release"

func main() {
	release.WriteProviderFiles()
	release.WriteNewVersion()
	release.CreateRelease()
}
