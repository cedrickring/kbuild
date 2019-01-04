workflow "Run all tests" {
  on = "push"
  resolves = ["Run all checks"]
}

action "Run all checks" {
  uses = "cedrickring/golang-action@master"
  args = "make"
}
