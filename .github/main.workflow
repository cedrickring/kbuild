workflow "Run all tests" {
  on = "push"
  resolves = ["cedrickring/golang-action@master"]
}

action "cedrickring/golang-action@master" {
  uses = "cedrickring/golang-action@master"
  args = "build -o test cmd/kbuild.go"
}
