workflow "Run all tests" {
  on = "push"
  resolves = ["Run all checks"]
}

action "Run all checks" {
  uses = "cedrickring/golang-action@1.0.0"
  args = "make"
}
