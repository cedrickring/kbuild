workflow "kbuild CI" {
  on = "push"
  resolves = ["Run all checks"]
}

action "Run all checks" {
  uses = "cedrickring/golang-action@1.1.0"
}

workflow "Upload all artifacts" {
  resolves = [
    "Upload darwin release",
    "Upload linux release",
    "Upload windows release",
    "docker://alpine",
  ]
  on = "push"
}

action "Build all binaries" {
  uses = "cedrickring/golang-action@1.1.0"
  args = "make build-all"
}

action "Upload darwin release" {
  uses = "cedrickring/upload-to-release@master"
  args = "kbuild_darwin_amd64 application/octet-stream"
  secrets = ["GITHUB_TOKEN"]
  needs = ["Build all binaries"]
  env = {
    WORKING_DIRECTORY = "out"
  }
}

action "Upload linux release" {
  uses = "cedrickring/upload-to-release@master"
  args = "kbuild_linux_amd64 application/octet-stream"
  secrets = ["GITHUB_TOKEN"]
  needs = ["Build all binaries"]
  env = {
    WORKING_DIRECTORY = "out"
  }
}

action "Upload windows release" {
  uses = "cedrickring/upload-to-release@master"
  args = "kbuild_windows_amd64.exe application/octet-stream"
  secrets = ["GITHUB_TOKEN"]
  needs = ["Build all binaries"]
  env = {
    WORKING_DIRECTORY = "out"
  }
}

action "docker://alpine" {
  uses = "docker://alpine"
  needs = ["Build all binaries"]
  args = "ls out"
}
