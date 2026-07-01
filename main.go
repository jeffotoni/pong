package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	paddleWidth  = 20
	paddleHeight = 100
	ballSize     = 20
	gameDuration = 40 * time.Second
	sampleRate   = 44100
)

var (
	//go:embed assets/PressStart2P-Regular.ttf
	pressStart2PTTF []byte
	//go:embed assets/pong_sound.wav
	pongSoundData []byte
	//go:embed assets/wall_sound.wav
	wallSoundData []byte

	largeFont font.Face
	smallFont font.Face
	tinyFont  font.Face
	audioCtx  *audio.Context
	hitSound  *audio.Player
	wallSound *audio.Player
)

type GameState int

const (
	StateEnterName GameState = iota
	StateRunning
	StatePaused
	StateGameOver
)

type Game struct {
	ballX, ballY              float64
	ballSpeedX, ballSpeedY    float64
	leftPaddleY, rightPaddleY float64
	leftScore, rightScore     int
	startTime                 time.Time
	state                     GameState
	computerErrorMargin       float64
	playerName                string
	nameInput                 []rune
	audioEnabled              bool
}

func (g *Game) initGame() {
	g.ballX = screenWidth / 2
	g.ballY = screenHeight / 2
	g.ballSpeedX = 8
	g.ballSpeedY = 8
	g.leftPaddleY = screenHeight / 2
	g.rightPaddleY = screenHeight / 2
	g.leftScore = 0
	g.rightScore = 0
	g.startTime = time.Now()
	g.state = StateEnterName
	g.computerErrorMargin = 2 // Computer is perfect initially
	g.playerName = ""
	g.nameInput = []rune{}
}

func (g *Game) Update() error {
	g.audioEnabled = audioCtx != nil && audioCtx.IsReady()

	if g.state != StateEnterName {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.initGame()
			return nil
		}
	}

	switch g.state {
	case StateEnterName:
		for _, c := range ebiten.InputChars() {
			if c == '\b' {
				if len(g.nameInput) > 2 {
					g.nameInput = g.nameInput[:len(g.nameInput)-1]
				}
			} else {
				g.nameInput = append(g.nameInput, c)
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyEnter) && len(g.nameInput) > 2 {
			g.playerName = string(g.nameInput)
			g.state = StateRunning
			g.startTime = time.Now()
		}
	case StatePaused:
		if ebiten.IsKeyPressed(ebiten.KeyP) {
			g.state = StateRunning
		}
	case StateGameOver:
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.initGame()
		}
	case StateRunning:
		if ebiten.IsKeyPressed(ebiten.KeyP) {
			g.state = StatePaused
		}

		// Move the ball
		g.ballX += g.ballSpeedX
		g.ballY += g.ballSpeedY

		// Bounce off top and bottom
		if g.ballY < 0 || g.ballY > screenHeight-ballSize {
			g.ballSpeedY *= -1
			playSound(wallSound, 1)
		}

		// Bounce off paddles
		if g.ballX < paddleWidth && g.ballY > g.leftPaddleY && g.ballY < g.leftPaddleY+paddleHeight {
			g.ballSpeedX *= -1
			playSound(hitSound, 1)
		}
		if g.ballX > screenWidth-paddleWidth-ballSize && g.ballY > g.rightPaddleY && g.ballY < g.rightPaddleY+paddleHeight {
			g.ballSpeedX *= -1
			playSound(hitSound, 1)
		}

		// Score points and reset ball position
		if g.ballX < 0 {
			g.rightScore++
			g.ballX, g.ballY = screenWidth/2, screenHeight/2
			showScore(g.playerName, g.rightScore, g.leftScore, 0)
		}
		if g.ballX > screenWidth {
			g.leftScore++
			g.ballX, g.ballY = screenWidth/2, screenHeight/2
			showScore(g.playerName, g.rightScore, g.leftScore, 0)
		}

		// Move left paddle (AI with error margin)
		if g.computerErrorMargin == 0 {
			// Perfect precision
			g.leftPaddleY = g.ballY - paddleHeight/2
		} else {
			// With error margin
			targetY := g.ballY + rand.Float64()*g.computerErrorMargin*2 - g.computerErrorMargin
			if targetY > g.leftPaddleY+paddleHeight/2 {
				g.leftPaddleY += 6 // Reduced speed for less precision
			} else if targetY < g.leftPaddleY+paddleHeight/2 {
				g.leftPaddleY -= 6 // Reduced speed for less precision
			}
		}

		// Move right paddle (Human)
		if ebiten.IsKeyPressed(ebiten.KeyUp) && g.rightPaddleY > 0 {
			g.rightPaddleY -= 10
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) && g.rightPaddleY < screenHeight-paddleHeight {
			g.rightPaddleY += 10
		}

		// pay := Payload{
		// 	Name:         g.playerName,
		// 	ScoreHumane:  g.rightScore,
		// 	ScoreMachine: g.leftScore,
		// }

		// bb, _ := json.Marshal(pay)
		// fmt.Printf("%s\n", string(bb))

		// Check game duration
		elapsed := time.Since(g.startTime)
		if elapsed >= gameDuration {
			showScore(g.playerName, g.rightScore, g.leftScore, 1)

			// Optional API post hooks can be added here.
			// go func(playerName string, rightScore, leftScore int) {
			// 	pay := Payload{
			// 		Name:         playerName,
			// 		ScoreHumane:  rightScore,
			// 		ScoreMachine: leftScore,
			// 	}

			// 	td := time.Duration(10 * time.Second)
			// 	url := ""

			// 	err := pay.Post(td, url)
			// 	if err != nil {
			// 		fmt.Printf("Player: %s, Score: %d - %d - %v\n", g.playerName, g.leftScore, g.rightScore, err)
			// 	} else {
			// 		fmt.Printf("Player: %s, Score: %d - %d - success\n", g.playerName, g.leftScore, g.rightScore)
			// 	}
			// }(g.playerName, g.rightScore, g.leftScore)

			g.state = StateGameOver
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	// Draw ball
	ebitenutil.DrawRect(screen, g.ballX, g.ballY, ballSize, ballSize, color.White)

	// Draw paddles
	ebitenutil.DrawRect(screen, 0, g.leftPaddleY, paddleWidth, paddleHeight, color.White)
	ebitenutil.DrawRect(screen, screenWidth-paddleWidth, g.rightPaddleY, paddleWidth, paddleHeight, color.White)

	// Draw center line
	for i := 0; i < screenHeight; i += 20 {
		ebitenutil.DrawRect(screen, screenWidth/2-1, float64(i), 2, 10, color.White)
	}

	// Draw scores
	text.Draw(screen, fmt.Sprintf("%d", g.leftScore), largeFont, screenWidth/2-150, 70, color.White)
	text.Draw(screen, fmt.Sprintf("%d", g.rightScore), largeFont, screenWidth/2+100, 70, color.White)

	// Draw timer
	elapsed := time.Since(g.startTime)
	remaining := gameDuration - elapsed
	if remaining < 0 {
		remaining = 0
	}
	text.Draw(screen, fmt.Sprintf("Time: %02d:%02d", int(remaining.Minutes()),
		int(remaining.Seconds())%60), largeFont, screenWidth/2-250, screenHeight-60, color.White)

	yellow := color.RGBA{255, 255, 0, 255}
	green := color.RGBA{50, 205, 50, 255}
	red := color.RGBA{255, 0, 0, 255}

	text.Draw(screen, "helloworldhub.io", tinyFont, screenWidth-874, screenHeight-49, yellow)
	text.Draw(screen, g.playerName, tinyFont, screenWidth-600, screenHeight-49, green)

	// Draw PAUSE and RESTART messages
	if g.state == StatePaused {
		msg := "PRESS 'P' GO BACK"
		text.Draw(screen, msg, smallFont, screenWidth-830, screenHeight-120, color.White)

		msg = "GAME OVER 'R'"
		text.Draw(screen, msg, smallFont, screenWidth-750, screenHeight-2, color.White)

	} else if g.state == StateGameOver {
		msg := "GAME OVER - 'R' TO RESTART"
		text.Draw(screen, msg, smallFont, screenWidth-900, screenHeight-350, red)
	} else {
		msg := "PRESS 'P' TO PAUSE"
		text.Draw(screen, msg, smallFont, screenWidth-830, screenHeight-120, color.White)

		msg = "GAME OVER 'R'"
		text.Draw(screen, msg, smallFont, screenWidth-750, screenHeight-2, color.White)
	}

	if g.state == StateEnterName {

		maxChars := 10
		displayName := string(g.nameInput)
		if len(displayName) > maxChars {
			displayName = displayName[:maxChars]
		}

		text.Draw(screen, "Enter Player Name", smallFont, screenWidth/2-150, screenHeight/2-50, yellow)
		text.Draw(screen, displayName, smallFont, screenWidth/2-100, screenHeight/2, green)
	}

	if audioCtx != nil && !g.audioEnabled {
		text.Draw(screen, "Audio locked by browser: click the page and press any key.", tinyFont, 20, screenHeight-20, color.White)
	}

	// if g.state == StatePaused {
	// 	text.Draw(screen, "PRESS 'P' TO RESUME", smallFont, screenWidth/2-100, screenHeight/2, color.White)
	// }
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func loadFont() {
	ttf, err := opentype.Parse(pressStart2PTTF)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72

	largeFont, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    48,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	smallFont, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	tinyFont, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    12,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func playSound(player *audio.Player, volume float64) {
	if player == nil || audioCtx == nil || !audioCtx.IsReady() {
		return
	}
	player.Rewind()
	player.SetVolume(volume)
	player.Play()
}

func loadSound() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("audio init panic: %v", r)
		}
	}()

	audioCtx = audio.NewContext(sampleRate)

	hitSoundDecoded, err := wav.Decode(audioCtx, bytes.NewReader(pongSoundData))
	if err != nil {
		return err
	}

	wallSoundDecoded, err := wav.Decode(audioCtx, bytes.NewReader(wallSoundData))
	if err != nil {
		return err
	}

	hitSound, err = audio.NewPlayer(audioCtx, hitSoundDecoded)
	if err != nil {
		return err
	}

	wallSound, err = audio.NewPlayer(audioCtx, wallSoundDecoded)
	if err != nil {
		return err
	}

	return nil
}

func shouldEnableAudio() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PONG_AUDIO"))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}

	return runtime.GOOS != "darwin"
}

func main() {
	loadFont()
	if shouldEnableAudio() {
		if err := loadSound(); err != nil {
			log.Printf("audio disabled: %v", err)
			audioCtx = nil
			hitSound = nil
			wallSound = nil
		}
	} else {
		log.Printf("audio disabled (set PONG_AUDIO=1 to force-enable)")
	}

	game := &Game{}
	game.initGame()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Pong in Go / @jeffotoni")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Payload struct {
	Name         string `json:"name"`
	ScoreHumane  int    `json:"score_humane"`
	ScoreMachine int    `json:"score_machine"`
	GameOver     int    `json:"game_over"`
}

func (p Payload) Post(td time.Duration, url string) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), td)
	defer cancel()
	err = sendPostRequest(ctx, url, p)
	return
}

func sendPostRequest(ctx context.Context, url string, payload Payload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %v", resp.Status)
	}

	return nil
}

func showScore(playerName string, rightScore, leftScore, gameOver int) {
	pay := Payload{
		Name:         playerName,
		ScoreHumane:  rightScore,
		ScoreMachine: leftScore,
		GameOver:     gameOver,
	}
	bb, _ := json.Marshal(pay)
	fmt.Printf("%s\n", string(bb))
}
