workflow "Run all tests" {
  on = "push"
  resolves = ["Run tests"]
}

action "Run tests" {
  uses = "docker://golang"
}
