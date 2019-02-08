workflow "kbuild CI" {
  on = "push"
  resolves = ["Run all checks"]
}

action "Run all checks" {
  uses = "cedrickring/golang-action@1.1.0"
}

workflow "Upload all artifacts" {
  resolves = ["Upload darwin release", "Upload linux release", "Upload windows release"]
  on = "push"
}

action "Build all binaries" {
  uses = "cedrickring/golang-action@1.1.0"
  args = "make build-all"
}

action "Test" {
  uses = "docker://alpine"
  needs = ["Build all binaries"]
  args = "ls -la"
}

action "Upload darwin release" {
  uses = "JasonEtco/upload-to-release@master"
  needs = ["Test"]
  args = "kbuild_darwin_amd64"
  secrets = ["GITHUB_TOKEN"]
}

action "Upload linux release" {
  uses = "JasonEtco/upload-to-release@master"
  needs = ["Test"]
  args = "kbuild_linux_amd64"
  secrets = ["GITHUB_TOKEN"]
}

action "Upload windows release" {
  uses = "JasonEtco/upload-to-release@master"
  needs = ["Test"]
  args = "kbuild_windows_amd64.exe"
  secrets = ["GITHUB_TOKEN"]
}
