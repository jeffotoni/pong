# Pong v7

[![GoDoc](https://godoc.org/github.com/jeffotoni/pong?status.svg)](https://godoc.org/github.com/jeffotoni/pong) [![Go Report](https://goreportcard.com/badge/github.com/jeffotoni/pong)](https://goreportcard.com/report/github.com/jeffotoni/pong) [![License](https://img.shields.io/github/license/jeffotoni/pong)](https://github.com/jeffotoni/pong/blob/main/LICENSE) ![GitHub last commit](https://img.shields.io/github/last-commit/jeffotoni/pong) ![GitHub contributors](https://img.shields.io/github/contributors/jeffotoni/pong) ![GitHub forks](https://img.shields.io/github/forks/jeffotoni/pong?style=social) ![GitHub stars](https://img.shields.io/github/stars/jeffotoni/pong?style=social)

A classic Pong game built with **Go** and **Ebitengine**, runnable on desktop and in the browser through **WebAssembly**.

This repository keeps only the final, clean version of the project while preserving the development journey as context.

---

## Pong in Action

![Pong preview](./pong.jpeg)

---

## About

Pong v7 focuses on the core gameplay loop:

- AI versus player paddle mode
- Real-time score and game timer
- Player name entry before match start
- Pause and restart controls
- Embedded font and sound assets
- Desktop and browser execution

---

## Requirements

- Go 1.22 or newer
- A modern browser with WebAssembly support for the web version

---

## Quick Start

```bash
git clone https://github.com/jeffotoni/pong.git
cd pong
go run .
```

---

## Controls

| Action | Key |
|---|---|
| Move player paddle up | Arrow Up |
| Move player paddle down | Arrow Down |
| Pause or resume | P |
| Restart match | R |
| Confirm name input | Enter |

---

## Run on Desktop

```bash
go run .
```

Build a local binary:

```bash
go build -o pong .
./pong
```

---

## Run in the Browser with WebAssembly

### Option A: Build with Go and serve static files

```bash
GOOS=js GOARCH=wasm go build -o web/pong.wasm .
cd web
python3 -m http.server 8080
```

Open:

```text
http://localhost:8080
```

### Option B: Use wasmserve

```bash
go run github.com/hajimehoshi/wasmserve@latest .
```

Open:

```text
http://localhost:8080/web/
```

Important:

- Run `wasmserve` from the project root.
- Do not run `wasmserve ./web`, because `web/` does not contain the Go entrypoint.
- `web/index.html` can load both `pong.wasm` and `/main.wasm`, so it works with both options above.

---

## Browser Audio

Browsers can block game audio until the page receives user interaction.

If the game starts without sound:

- Click the game page or canvas once.
- Press any key.
- The in-game message `Audio locked by browser...` disappears when audio is ready.

---

## Desktop Audio (macOS)

On macOS, audio is disabled by default to avoid CoreAudio startup failures in some environments.

To force-enable audio:

```bash
PONG_AUDIO=1 go run .
```

---

## Project Layout

```text
.
├── assets/
│   ├── PressStart2P-Regular.ttf
│   ├── pong_sound.wav
│   └── wall_sound.wav
├── web/
│   ├── index.html
│   └── wasm_exec.js
├── main.go
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

---

## Troubleshooting

### `expected magic word ... found 34 30 34 20`

This means the browser received a `404` HTML response instead of a real `.wasm` file.

Fixes:

- With Option A, make sure `web/pong.wasm` exists after running the build command.
- With Option B, run `go run github.com/hajimehoshi/wasmserve@latest .` from the project root and open `/web/`.

### No sound in the browser

This is usually caused by browser autoplay policy.

Fix:

- Click the page and press a key once to unlock audio.

---

## Development Journey

| Version | Focus |
|---|---|
| Version 0 | Basic board and paddle rendering |
| Version 1 | Ball movement and collisions |
| Version 2 | Scoring and reset logic |
| Version 3 | Timer and round flow |
| Version 4 | Name input and UI messages |
| Version 5 | AI paddle behavior tuning |
| Version 6 | Sound integration |
| Version 7 | Current clean version with desktop and web structure |

---

## Contributing

Contributions are welcome.

Suggested ways to improve the project:

- Add tests for scoring and collision behavior.
- Improve AI difficulty profiles.
- Add visual themes and additional sound packs.
- Open an issue with ideas or bugs.

---

## License

This project is open source under the **MIT License**.

Copyright (c) 2026 Jefferson Otoni Lima.

See [LICENSE](./LICENSE) for the full license text.
