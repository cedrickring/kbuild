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

action "cd into out directory" {
  uses = "docker://alpine"
  needs = ["Build all binaries"]
  args = "cd out"
  runs = "/bin/sh -c"
}

action "Upload darwin release" {
  uses = "JasonEtco/upload-to-release@master"
  args = "kbuild_darwin_amd64"
  secrets = ["GITHUB_TOKEN"]
  needs = ["cd into out directory"]
}

action "Upload linux release" {
  uses = "JasonEtco/upload-to-release@master"
  args = "kbuild_linux_amd64"
  secrets = ["GITHUB_TOKEN"]
  needs = ["cd into out directory"]
}

action "Upload windows release" {
  uses = "JasonEtco/upload-to-release@master"
  args = "kbuild_windows_amd64.exe"
  secrets = ["GITHUB_TOKEN"]
  needs = ["cd into out directory"]
}
