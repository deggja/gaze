<div align="center">
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-v1.21-brightgreen.svg" alt="go version">
  </a>
</div>

<div align="center">

  <h1>gaze</h1>
  <h3>look at k8s steadily and intently, especially in admiration, surprise, or thought</h3>

</div>

## Contents
- [What is Gaze?](#what-is-gaze)
- [Features](#features)
- [Installation](#installation)
  - [Homebrew](#homebrew-)
  - [Build from source](#build-from-source-)
- [Usage](#usage)
  - [Starting the app](#starting-the-app)
- [Kubernetes interaction](#kubernetes-interaction)
- [Contribute](#contribute-)
- [Acknowledgments](#acknowledgments)
- [License](#license)

## ⭐ What is Gaze? ⭐

Gaze provides a TUI that lets you quickly search through your Kubernetes cluster for keywords in application logs.

## Features

- Search Kubernetes pod logs for a comprehensive list of error-related keywords.
- Interactive selection of pods to inspect logs in detail.
- Highlighting of keywords within the log output for easy identification.

## Installation

### Homebrew 🍺

To install gaze using Homebrew, you can run the following commands:

```sh
brew tap deggja/gaze https://github.com/deggja/gaze
brew install gaze
```

### Build from source 💻
```
git clone https://github.com/deggja/gaze.git --depth 1
cd gaze
go build -o gaze
```

## Usage

### Starting the app

To start the game, simply run the compiled binary:

```sh
./gaze
```

## Kubernetes interaction

Gaze will require access to your Kubernetes cluster. Ensure your `kubeconfig` is set up correctly before starting the game. The application currently expects the kubeconfig at its default location.

## Contribute 🔨

Feel free to dive in! [Open an issue](https://github.com/deggja/gaze/issues) or submit PRs.

## Acknowledgments

Gaze uses [bubbletea](https://github.com/charmbracelet/bubbletea).

## License

Serpent is released under the MIT License. Check out the [LICENSE](https://github.com/deggja/gaze/LICENSE) file for more information.