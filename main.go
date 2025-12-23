package main

import (
	"cart/w4"
	"math"
	"math/rand"
	"time"
)

const (
	playerSpeed            = 0.75
	stoneSize              = 2
	stoneSpeed             = 3.5
	zombieSize             = 8
	zombieSpacing          = 2.0
	poisonSpeed            = 1.0
	poisonSize             = 8
	poisonDuration         = 40
	stateMenu              = 0
	statePlaying           = 1
	stateGameOver          = 2
	stateTransitionToFinal = 3
	stateFinalStage        = 4
	statePaused            = 5
	mapWidth               = 200
	mapHeight              = 200
	numScenery             = 150
	totalKeys              = 3
	maxPlayerHealth        = 3
	invincibilityDuration  = 90
	poisonUnlockItemCount  = 1
	messageDisplayTime     = 120
	bossFightAmmo          = 3
	bossFightReloadTime    = 90
	gravity                = 0.15
	jumpForce              = -3.5
	transitionDuration     = 180
	totalWaves             = 3
	baseNumZombies         = 5
	zombiesPerWave         = 3
	maxZombies             = baseNumZombies + (zombiesPerWave * (totalWaves - 1))
	bossMaxHealth          = 12
	bossWidth              = 16
	bossHeight             = 16
	numFloorDetails        = 200
)

// Sprites - Constantes e Arrays
const (
	playerFlags        = w4.BLIT_2BPP
	playerSpriteWidth  = 8
	playerSpriteHeight = 8
	playerSpriteFlags  = 0
	bossFlags          = w4.BLIT_2BPP
	heartWidth         = 8
	heartHeight        = 8
	heartFlags         = w4.BLIT_2BPP
	keyWidth           = 8
	keyHeight          = 8
	keyFlags           = w4.BLIT_1BPP
	poisonWidth        = 8
	poisonHeight       = 8
	poisonFlags        = w4.BLIT_1BPP
)

var playerSprite = [16]byte{0xd5, 0x57, 0x55, 0x55, 0x50, 0x05, 0x40, 0x01, 0x68, 0x29, 0x68, 0x29, 0x40, 0x01, 0x55, 0x55}
var bossSprite = [64]byte{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xa5, 0x55, 0xaa, 0xaa, 0xa4, 0x04, 0x6a, 0xaa, 0x50, 0x04, 0x1a,
	0xa5, 0x40, 0x10, 0x1a, 0x90, 0x00, 0x00, 0x06, 0x90, 0x50, 0x05, 0x06, 0x90, 0x54, 0x15, 0x06, 0x90, 0x14, 0x14,
	0x06, 0x90, 0x00, 0x80, 0x06, 0xa4, 0x02, 0x80, 0x1a, 0xa9, 0x40, 0x01, 0x6a, 0xaa, 0x44, 0x11, 0xaa, 0xaa, 0x44,
	0x11, 0xaa, 0xaa, 0x55, 0x55, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa}
var zombieSprite = [16]byte{0xaa, 0xaa, 0xaa, 0x5a, 0x95, 0x55, 0x55, 0x55, 0x41, 0x41, 0x55, 0x55, 0x40, 0x01, 0x55, 0x55}
var heart = [16]byte{0xaa, 0xaa, 0xbe, 0xbe, 0xd7, 0xc3, 0xd5, 0x53, 0xd5, 0x57, 0xb5, 0x5e, 0xad, 0x7a, 0xab, 0xea}
var keySprite = [8]byte{0xC3, 0xBD, 0xBD, 0xC3, 0xE7, 0xE1, 0xE7, 0xE1}
var poisonSprite = [8]byte{0x3C, 0x42, 0x81, 0x81, 0x81, 0x81, 0x42, 0x3C}

type Stone struct {
	x, y, vx, vy float64
	active       bool
}
type Poison struct {
	x, y, vx, vy float64
	active       bool
	timer        int
}
type Zombie struct {
	x, y  float64
	alive bool
}
type Boss struct {
	x, y  float64
	alive bool
	hp    int
}
type SceneryObject struct{ x, y int }
type Key struct {
	x, y  float64
	found bool
}
type PoisonItem struct {
	x, y  float64
	found bool
}

var (
	playerX, playerY    float64
	playerVY            float64 = 0
	playerHealth                = maxPlayerHealth
	ammoLeft                    = pistolAmmo
	zombieSpeed         float64 = 0.3
	reloadTime          int     = 60
	pistolAmmo          int     = 3
	isGrounded          bool    = false
	reloading                   = false
	reloadTimer                 = 0
	isPlayerInvincible  bool    = false
	invincibilityTimer  int     = 0
	hasPoisonPower      bool    = false
	lastDirX, lastDirY  float64 = 1, 0
	cameraX, cameraY    float64
	frameCounterForSeed int64 = 0
	gameState           int   = 0
	musicTimer          int   = 0
	transitionTimer     int   = 0
	currentWave         int   = 1
	currentNumZombies   int
	keysCollected       int
	message             string
	messageTimer        int
	previousGameState   int
	pauseSelection      int = 0
	prevMouseState      uint8
	prevGamepadState    uint8
	gamepadJustPressed  uint8
	mouseJustPressed    uint8
	stone               Stone
	poison              Poison
	boss                Boss
	zombies             [maxZombies]Zombie
	scenery             [numScenery]SceneryObject
	keys                [totalKeys]Key
	poisonItems         [poisonUnlockItemCount]PoisonItem
)

//go:export start
func start() {
	w4.PALETTE[0] = 0xF5F2D0
	w4.PALETTE[1] = 0x306850
	w4.PALETTE[2] = 0x1c1916
	w4.PALETTE[3] = 0x0d6da1
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < numScenery; i++ {
		scenery[i] = SceneryObject{x: rand.Intn(mapWidth), y: rand.Intn(mapHeight)}
	}
}

//go:export update
func update() {
	frameCounterForSeed++

	gamepad := *w4.GAMEPAD1
	gamepadJustPressed = gamepad & (gamepad ^ prevGamepadState)
	prevGamepadState = gamepad

	mouse := *w4.MOUSE_BUTTONS
	mouseJustPressed = mouse & (mouse ^ prevMouseState)
	prevMouseState = mouse

	if (gameState == statePlaying || gameState == stateFinalStage) && (mouseJustPressed&w4.MOUSE_LEFT != 0) {
		previousGameState = gameState
		gameState = statePaused
		pauseSelection = 0
		w4.Tone(262, 10, 80, w4.TONE_PULSE1)
		return
	}

	switch gameState {
	case stateMenu:
		playIntroMusic()
		*w4.DRAW_COLORS = 4
		w4.Text("MONSTER SURVIVAL", 20, 25)
		w4.Text("CONTROLS:", 10, 50)
		w4.Text("\x86 Up", 10, 60)
		w4.Text("\x87 Down", 10, 70)
		w4.Text("\x84 Left", 10, 80)
		w4.Text("\x85 Right", 10, 90)
		w4.Text("\x81 Z - Poison orb", 10, 100)
		w4.Text("\x80 X - Throw stone", 10, 110)
		w4.Text("L-CLICK - Pause", 10, 120)
		w4.Text("Collect 3 keys", 1, 135)
		w4.Text("and kill all zombies", 1, 145)

		if gamepadJustPressed != 0 {
			rand.Seed(frameCounterForSeed)
			gameState = statePlaying
			initGame()
		}

	case statePlaying:
		playDarkMusic()
		drawFloorPattern()

		*w4.DRAW_COLORS = 1
		for _, obj := range scenery {
			w4.Rect(obj.x-int(cameraX), obj.y-int(cameraY), 1, 1)
		}

		if isPlayerInvincible {
			invincibilityTimer--
			if invincibilityTimer <= 0 {
				isPlayerInvincible = false
			}
		}

		dx, dy := 0.0, 0.0
		if gamepad&w4.BUTTON_LEFT != 0 {
			dx -= 1
		}
		if gamepad&w4.BUTTON_RIGHT != 0 {
			dx += 1
		}
		if gamepad&w4.BUTTON_UP != 0 {
			dy -= 1
		}
		if gamepad&w4.BUTTON_DOWN != 0 {
			dy += 1
		}
		length := math.Hypot(dx, dy)
		if length != 0 {
			dx /= length
			dy /= length
			playerX += dx * playerSpeed
			playerY += dy * playerSpeed
			lastDirX = dx
			lastDirY = dy
		} else {
			if lastDirX == 0 && lastDirY == 0 {
				lastDirX = 1
			}
		}

		playerX = math.Max(0, math.Min(float64(mapWidth-8), playerX))
		playerY = math.Max(12, math.Min(float64(mapHeight-8), playerY))
		cameraX = playerX - 80 + 4
		cameraY = playerY - 80 + 4
		cameraX = math.Max(0, math.Min(float64(mapWidth-160), cameraX))
		cameraY = math.Max(0, math.Min(float64(mapHeight-160), cameraY))

		if reloading {
			reloadTimer++
			if reloadTimer >= reloadTime {
				ammoLeft = pistolAmmo
				reloading = false
			}
		}

		handlePlayerAction(gamepad)

		if stone.active {
			stone.x += stone.vx
			stone.y += stone.vy
			checkStoneHit(int(stone.x), int(stone.y))
			if stone.x < 0 || stone.x > mapWidth-stoneSize || stone.y < 0 || stone.y > mapHeight-stoneSize {
				stone.active = false
			} else {
				*w4.DRAW_COLORS = 3
				w4.Rect(int(stone.x-cameraX), int(stone.y-cameraY), stoneSize, stoneSize)
			}
		}
		if poison.active {
			poison.x += poison.vx
			poison.y += poison.vy
			checkPoisonHit(int(poison.x), int(poison.y))
			poison.timer--
			if poison.timer <= 0 || poison.x < 0 || poison.x > mapWidth || poison.y < 0 || poison.y > mapHeight {
				poison.active = false
			} else {
				*w4.DRAW_COLORS = 0x4321
				drawFilledCircle(int(poison.x-cameraX)+poisonSize/2, int(poison.y-cameraY)+poisonSize/2, poisonSize/2)
			}
		}

		moveZombiesTowardsPlayer()
		checkZombieCollision()
		checkKeyCollection()
		checkPoisonItemCollection()

		allZombiesDead := true
		for i := 0; i < currentNumZombies; i++ {
			if zombies[i].alive {
				allZombiesDead = false
				break
			}
		}

		if allZombiesDead && keysCollected == totalKeys {
			if currentWave < totalWaves {
				currentWave++
				setupWave()
			} else {
				gameState = stateTransitionToFinal
				transitionTimer = 0
			}
		}

		*w4.DRAW_COLORS = 0x0C
		for _, k := range keys {
			if !k.found {
				w4.Blit(&keySprite[0], int(k.x-cameraX), int(k.y-cameraY), keyWidth, keyHeight, keyFlags)
			}
		}

		*w4.DRAW_COLORS = 0x21
		for _, f := range poisonItems {
			if !f.found {
				w4.Blit(&poisonSprite[0], int(f.x-cameraX), int(f.y-cameraY), poisonWidth, poisonHeight, poisonFlags)
			}
		}

		*w4.DRAW_COLORS = 0x3320
		for i := 0; i < currentNumZombies; i++ {
			z := &zombies[i]
			if z.alive {
				w4.Blit(&zombieSprite[0], int(z.x-cameraX), int(z.y-cameraY), 8, 8, w4.BLIT_2BPP)
			}
		}

		shouldDrawPlayer := true
		if isPlayerInvincible {
			if invincibilityTimer%8 < 4 {
				shouldDrawPlayer = false
			}
		}
		if shouldDrawPlayer {
			*w4.DRAW_COLORS = 0x0341
			w4.Blit(&playerSprite[0], int(playerX-cameraX), int(playerY-cameraY), 8, 8, w4.BLIT_2BPP)
		}

		drawReloadBar()
		if messageTimer > 0 {
			*w4.DRAW_COLORS = 3
			w4.Text(message, (160-len(message)*8)/2, 145)
			messageTimer--
		}
		drawHUD()

	case statePaused:
		updatePaused()

	case stateGameOver:
		playGameOverMusic()
		*w4.DRAW_COLORS = 3
		if playerHealth <= 0 {
			w4.Text("GAME OVER", 50, 70)
		} else {
			w4.Text("YOU WON!", 45, 70)
		}
		w4.Text("Press Z to restart", 12, 90)
		if gamepad&w4.BUTTON_2 != 0 {
			rand.Seed(frameCounterForSeed)
			gameState = stateMenu
			restartGame()
		}

	case stateTransitionToFinal:
		switch transitionTimer {
		case 10:
			w4.Tone(180, 10, 100, w4.TONE_PULSE2)
		case 30:
			w4.Tone(200, 10, 100, w4.TONE_PULSE2)
		case 55:
			startFreq := uint(220)
			endFreq := uint(150)
			packedFreq := startFreq | (endFreq << 16)
			w4.Tone(packedFreq, 45, 100, w4.TONE_PULSE2)
		}

		*w4.DRAW_COLORS = 3
		w4.Rect(0, 0, 160, 160)

		transitionTimer++
		if transitionTimer >= transitionDuration {
			gameState = stateFinalStage
			initFinalStage()
		}

	case stateFinalStage:
		playBossMusic()

		*w4.DRAW_COLORS = 2
		w4.Rect(0, 150, 160, 10)

		if gamepad&w4.BUTTON_LEFT != 0 {
			playerX -= playerSpeed
			lastDirX = -1
		}
		if gamepad&w4.BUTTON_RIGHT != 0 {
			playerX += playerSpeed
			lastDirX = 1
		}

		playerVY += gravity
		playerY += playerVY
		if playerY >= 150-8 {
			playerY = 150 - 8
			playerVY = 0
			isGrounded = true
		} else {
			isGrounded = false
		}

		if gamepad&w4.BUTTON_UP != 0 && isGrounded {
			playerVY = jumpForce
			isGrounded = false
		}
		playerX = math.Max(0, math.Min(160-8, playerX))

		if reloading {
			reloadTimer++
			if reloadTimer >= reloadTime {
				ammoLeft = bossFightAmmo
				reloading = false
				reloadTimer = 0
			}
		}

		if gamepad&w4.BUTTON_1 != 0 && !stone.active && !reloading && ammoLeft > 0 {
			shootStoneFinalStage()
			ammoLeft--
			if ammoLeft <= 0 {
				reloading = true
				reloadTimer = 0
			}
		}

		if stone.active {
			stone.x += stone.vx
			stone.y += stone.vy
			checkStoneHitBoss()
			if stone.x < 0 || stone.x > 160 {
				stone.active = false
			} else {
				*w4.DRAW_COLORS = 3
				w4.Rect(int(stone.x), int(stone.y), stoneSize, stoneSize)
			}
		}

		moveBossFinalStage()
		checkBossCollision()
		if isPlayerInvincible {
			invincibilityTimer--
			if invincibilityTimer <= 0 {
				isPlayerInvincible = false
			}
		}

		if boss.alive {
			*w4.DRAW_COLORS = 0x4021
			w4.Blit(&bossSprite[0], int(boss.x), int(boss.y), bossWidth, bossHeight, bossFlags)
		}

		shouldDrawPlayer := true
		if isPlayerInvincible {
			if invincibilityTimer%8 < 4 {
				shouldDrawPlayer = false
			}
		}
		if shouldDrawPlayer {
			*w4.DRAW_COLORS = 0x0341
			w4.Blit(&playerSprite[0], int(playerX), int(playerY), 8, 8, w4.BLIT_2BPP)
		}

		if messageTimer > 0 {
			*w4.DRAW_COLORS = 4
			w4.Text(message, (160-len(message)*8)/2, 20)
			messageTimer--
		}

		drawReloadBar()
		drawBossHealthBar()
		drawHUD()
	}
}

// =============================================================================
// GERENCIAMENTO DE ESTADO DO JOGO
// =============================================================================

func initGame() {
	playerHealth = maxPlayerHealth
	ammoLeft = pistolAmmo
	reloading = false
	reloadTimer = 0
	isPlayerInvincible = false
	invincibilityTimer = 0
	hasPoisonPower = false
	musicTimer = 0
	currentWave = 1
	lastDirX, lastDirY = 1, 0
	playerX, playerY = mapWidth/2, mapHeight/2

	for i := 0; i < poisonUnlockItemCount; i++ {
		itemX := float64(rand.Intn(mapWidth - 8))
		itemY := float64(rand.Intn((mapHeight-8)-12) + 12)
		poisonItems[i] = PoisonItem{x: itemX, y: itemY, found: false}
	}
	setupWave()
}

func restartGame() {
	initGame()
}

func setupWave() {
	currentNumZombies = baseNumZombies + (currentWave-1)*zombiesPerWave
	if currentNumZombies > maxZombies {
		currentNumZombies = maxZombies
	}
	zombieSpeed = 0.3 + (0.07 * float64(currentWave-1))
	reloadTime = 60 + (20 * (currentWave - 1))
	pistolAmmo = 3 - (currentWave - 1)
	if pistolAmmo < 1 {
		pistolAmmo = 1
	}
	playerX, playerY = mapWidth/2, mapHeight/2
	poison.active = false
	stone.active = false
	keysCollected = 0
	ammoLeft = pistolAmmo
	reloading = false
	spawnZombies()

	for i := 0; i < totalKeys; i++ {
		keyX := float64(rand.Intn(mapWidth - 8))
		keyY := float64(rand.Intn((mapHeight-8)-12) + 12)
		keys[i] = Key{x: keyX, y: keyY, found: false}
	}
	if currentWave > 1 {
		message = "NEW WAVE ARRIVING!"
		messageTimer = messageDisplayTime
	}
}

func initFinalStage() {
	playerX = 10
	playerY = 150 - 8
	playerVY = 0
	isGrounded = true
	playerHealth = maxPlayerHealth
	isPlayerInvincible = false
	invincibilityTimer = 0
	musicTimer = 0
	messageTimer = messageDisplayTime
	stone.active = false
	lastDirX = 1
	lastDirY = 0

	ammoLeft = bossFightAmmo
	reloading = false
	reloadTimer = 0
	reloadTime = bossFightReloadTime

	boss.alive = true
	boss.hp = bossMaxHealth
	boss.x = 160 - float64(bossWidth) - 10
	boss.y = 150 - float64(bossHeight) + 3
}

// =============================================================================
// FUNÇÕES DE ÁUDIO
// =============================================================================

func playIntroMusic() {
	switch musicTimer {
	case 0:
		w4.Tone(659, 10, 100, w4.TONE_PULSE1)
	case 15:
		w4.Tone(587, 10, 100, w4.TONE_PULSE1)
	case 30:
		w4.Tone(659, 10, 100, w4.TONE_PULSE1)
	case 45:
		w4.Tone(784, 20, 100, w4.TONE_PULSE1)
	case 60:
		w4.Tone(880, 8, 100, w4.TONE_PULSE1)
	case 75:
		w4.Tone(784, 8, 100, w4.TONE_PULSE1)
	}
	musicTimer++
	if musicTimer > 90 {
		musicTimer = 0
	}
}

func playDarkMusic() {
	switch musicTimer {
	case 0:
		w4.Tone(110, 40, 100, w4.TONE_PULSE2)
	case 120:
		w4.Tone(123, 40, 100, w4.TONE_PULSE2)
	}
	musicTimer++
	if musicTimer > 240 {
		musicTimer = 0
	}
}

func playGameOverMusic() {
	switch musicTimer {
	case 0:
		w4.Tone(196, 15, 50, w4.TONE_PULSE1)
	case 20:
		w4.Tone(164, 15, 50, w4.TONE_PULSE1)
	case 40:
		w4.Tone(146, 15, 50, w4.TONE_PULSE1)
	}
	musicTimer++
	if musicTimer > 60 {
		musicTimer = 61
	}
}

func playBossMusic() {
	switch musicTimer {
	case 0:
		w4.Tone(65, 60, 100, w4.TONE_PULSE2)
	case 70:
		w4.Tone(92, 40, 100, w4.TONE_PULSE2)
	case 120:
		w4.Tone(261, 5, 80, w4.TONE_PULSE1)
	case 128:
		w4.Tone(261, 5, 80, w4.TONE_PULSE1)
	}
	musicTimer++
	if musicTimer > 180 {
		musicTimer = 0
	}
}

func drawHUD() {
	hudY := 0
	*w4.DRAW_COLORS = 3
	w4.Rect(0, hudY, 160, 12)

	contentY := hudY + 2
	*w4.DRAW_COLORS = 4

	w4.Text("HP:", 5, contentY)
	*w4.DRAW_COLORS = 0x0341
	for i := 0; i < playerHealth; i++ {
		w4.Blit(&heart[0], 28+i*10, contentY, heartWidth, heartHeight, heartFlags)
	}
	*w4.DRAW_COLORS = 4

	if gameState != stateFinalStage {
		w4.Text("KEYS:", 80, contentY)
		*w4.DRAW_COLORS = 0x0C
		for i := 0; i < keysCollected; i++ {
			w4.Blit(&keySprite[0], 120+i*10, contentY, keyWidth, keyHeight, keyFlags)
		}
	} else {
		w4.Text("AMMO: ", 80, contentY)
		if reloading {
			*w4.DRAW_COLORS = 2
		} else {
			*w4.DRAW_COLORS = 4
			for i := 0; i < ammoLeft; i++ {
				w4.Rect(115+i*6, contentY, 4, 6)
			}
		}
	}
}

func drawFloorPattern() {
	*w4.DRAW_COLORS = 1
	w4.Rect(0, 0, 160, 160)

	centerX := int(float64(mapWidth)/2 - cameraX)
	centerY := int(float64(mapHeight)/2 - cameraY)

	innerRadius := 40
	outerRadius := int(float64(innerRadius) * 1.5)

	*w4.DRAW_COLORS = 2
	drawCircle(centerX, centerY, innerRadius)
	drawCircle(centerX, centerY, outerRadius)
}

func drawCircle(x0, y0, r int) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	for x >= y {
		w4.Rect(x0+x, y0+y, 1, 1)
		w4.Rect(x0+y, y0+x, 1, 1)
		w4.Rect(x0-y, y0+x, 1, 1)
		w4.Rect(x0-x, y0+y, 1, 1)
		w4.Rect(x0-x, y0-y, 1, 1)
		w4.Rect(x0-y, y0-x, 1, 1)
		w4.Rect(x0+y, y0-x, 1, 1)
		w4.Rect(x0+x, y0-y, 1, 1)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}
}

func drawFilledCircle(x0, y0, r int) {
	for y := -r; y <= r; y++ {
		halfWidth := int(math.Sqrt(float64(r*r - y*y)))
		w4.Rect(x0-halfWidth, y0+y, uint(halfWidth*2+1), 1)
	}
}

func drawReloadBar() {
	if !reloading {
		return
	}
	const maxBarWidth = 8.0
	progress := float64(reloadTimer) / float64(reloadTime)
	barWidth := uint(progress * maxBarWidth)

	var barX, barY int
	if gameState == stateFinalStage {
		barX = int(playerX)
		barY = int(playerY) + 10
	} else {
		barX = int(playerX - cameraX)
		barY = int(playerY-cameraY) + 10
	}

	*w4.DRAW_COLORS = 2
	w4.Rect(barX, barY, uint(maxBarWidth), 2)
	if barWidth > 0 {
		*w4.DRAW_COLORS = 4
		w4.Rect(barX, barY, barWidth, 2)
	}
}

func drawBossHealthBar() {
	if !boss.alive {
		return
	}
	const maxBarWidth = 140.0
	barX := (160 - int(maxBarWidth)) / 2
	barY := 15
	progress := float64(boss.hp) / float64(bossMaxHealth)
	if progress < 0 {
		progress = 0
	}
	barWidth := uint(progress * maxBarWidth)
	*w4.DRAW_COLORS = 2
	w4.Rect(barX-1, barY-1, uint(maxBarWidth)+2, 10)
	if barWidth > 0 {
		*w4.DRAW_COLORS = 4
		w4.Rect(barX, barY, barWidth, 8)
	}
}

func updatePaused() {
	if gamepadJustPressed&w4.BUTTON_UP != 0 {
		pauseSelection--
		w4.Tone(196, 5, 80, w4.TONE_TRIANGLE)
	}
	if gamepadJustPressed&w4.BUTTON_DOWN != 0 {
		pauseSelection++
		w4.Tone(196, 5, 80, w4.TONE_TRIANGLE)
	}

	if pauseSelection < 0 {
		pauseSelection = 2
	}
	if pauseSelection > 2 {
		pauseSelection = 0
	}

	if gamepadJustPressed&w4.BUTTON_1 != 0 || mouseJustPressed&w4.MOUSE_LEFT != 0 {
		w4.Tone(330, 10, 80, w4.TONE_PULSE1)
		switch pauseSelection {
		case 0:
			gameState = previousGameState
		case 1:
			initGame()
			gameState = statePlaying
		case 2:
			gameState = stateMenu
		}
		return
	}

	*w4.DRAW_COLORS = 0x31
	w4.Rect(30, 50, 100, 65)

	*w4.DRAW_COLORS = 4
	w4.Text("PAUSED", 57, 55)

	*w4.DRAW_COLORS = 3
	w4.Text("Continue", 50, 75)
	w4.Text("Restart", 50, 85)
	w4.Text("Main Menu", 50, 95)

	*w4.DRAW_COLORS = 4
	w4.Text(">", 40, 75+(pauseSelection*10))
}

func handlePlayerAction(gamepad uint8) {
	if gamepad&w4.BUTTON_2 != 0 && hasPoisonPower && !poison.active && (lastDirX != 0 || lastDirY != 0) {
		castPoison()
	}
	if gamepad&w4.BUTTON_1 != 0 && !stone.active && ammoLeft > 0 && !reloading && (lastDirX != 0 || lastDirY != 0) {
		shootStone()
		ammoLeft--
		if ammoLeft == 0 {
			reloading = true
			reloadTimer = 0
		}
	}
}

func shootStone() {
	stone.active = true
	stone.x = playerX + 4 - stoneSize/2
	stone.y = playerY + 4 - stoneSize/2
	stone.vx = lastDirX * stoneSpeed
	stone.vy = lastDirY * stoneSpeed
}

func shootStoneFinalStage() {
	stone.active = true
	stone.x = playerX + 4 - stoneSize/2
	stone.y = playerY + 4 - stoneSize/2
	stone.vx = lastDirX * stoneSpeed
	stone.vy = 0
}

func castPoison() {
	poison.active = true
	poison.x = playerX + 4 - poisonSize/2
	poison.y = playerY + 4 - poisonSize/2
	poison.vx = lastDirX * poisonSpeed
	poison.vy = lastDirY * poisonSpeed
	poison.timer = poisonDuration
}

func spawnZombies() {
	for i := 0; i < currentNumZombies; i++ {
		for {
			x := float64(rand.Intn(mapWidth - zombieSize))
			y := float64(rand.Intn(mapHeight-zombieSize-12) + 12)
			if math.Hypot(x-playerX, y-playerY) > 100 {
				zombies[i] = Zombie{x: x, y: y, alive: true}
				break
			}
		}
	}
	for i := currentNumZombies; i < maxZombies; i++ {
		zombies[i].alive = false
	}
}

func moveZombiesTowardsPlayer() {
	for i := 0; i < currentNumZombies; i++ {
		z := &zombies[i]
		if !z.alive {
			continue
		}
		dx := playerX - z.x
		dy := playerY - z.y
		dist := math.Hypot(dx, dy)
		if dist > 0 {
			z.x += (dx / dist) * zombieSpeed
			z.y += (dy / dist) * zombieSpeed
		}
	}
	for i := 0; i < currentNumZombies; i++ {
		z1 := &zombies[i]
		if !z1.alive {
			continue
		}
		for j := i + 1; j < currentNumZombies; j++ {
			z2 := &zombies[j]
			if !z2.alive {
				continue
			}
			dx := z1.x - z2.x
			dy := z1.y - z2.y
			dist := math.Hypot(dx, dy)
			targetDist := float64(zombieSize) + zombieSpacing
			if dist < targetDist {
				overlap := (targetDist - dist) / 2
				if dist == 0 {
					dx = 1
					dy = 0
				} else {
					dx /= dist
					dy /= dist
				}
				z1.x += dx * overlap
				z1.y += dy * overlap
				z2.x -= dx * overlap
				z2.y -= dy * overlap
			}
		}
	}
}

func moveBossFinalStage() {
	if !boss.alive {
		return
	}
	const bossSpeed = 0.2
	if boss.x+float64(bossWidth)/2 < playerX+4 {
		boss.x += bossSpeed
	}
	if boss.x+float64(bossWidth)/2 > playerX+4 {
		boss.x -= bossSpeed
	}
	boss.x = math.Max(0, math.Min(160-float64(bossWidth), boss.x))
}

func checkStoneHit(px, py int) {
	for i := 0; i < currentNumZombies; i++ {
		z := &zombies[i]
		if z.alive && px < int(z.x)+zombieSize && px+stoneSize > int(z.x) && py < int(z.y)+zombieSize && py+stoneSize > int(z.y) {
			z.alive = false
			stone.active = false
			break
		}
	}
}

func checkPoisonHit(fx, fy int) {
	for i := 0; i < currentNumZombies; i++ {
		z := &zombies[i]
		if z.alive && fx < int(z.x)+zombieSize && fx+poisonSize > int(z.x) && fy < int(z.y)+zombieSize && fy+poisonSize > int(z.y) {
			z.alive = false
		}
	}
}

func checkZombieCollision() {
	if isPlayerInvincible {
		return
	}
	for i := 0; i < currentNumZombies; i++ {
		z := &zombies[i]
		if z.alive && int(playerX) < int(z.x)+zombieSize && int(playerX)+8 > int(z.x) && int(playerY) < int(z.y)+zombieSize && int(playerY)+8 > int(z.y) {
			playerHealth--
			isPlayerInvincible = true
			invincibilityTimer = invincibilityDuration
			if playerHealth <= 0 {
				gameState = stateGameOver
			}
			z.x -= (playerX - z.x) * 0.1
			z.y -= (playerY - z.y) * 0.1
			break
		}
	}
}

func checkStoneHitBoss() {
	if !stone.active || !boss.alive {
		return
	}
	if stone.x < boss.x+float64(bossWidth) && stone.x+stoneSize > boss.x && stone.y < boss.y+float64(bossHeight) && stone.y+stoneSize > boss.y {
		boss.hp--
		stone.active = false
		w4.Tone(120, 5, 80, w4.TONE_NOISE)
		if boss.hp <= 0 {
			boss.alive = false
			playerHealth = maxPlayerHealth
			gameState = stateGameOver
		}
	}
}

func checkBossCollision() {
	if isPlayerInvincible || !boss.alive {
		return
	}
	if playerX < boss.x+float64(bossWidth) && playerX+8 > boss.x && playerY < boss.y+float64(bossHeight) && playerY+8 > boss.y {
		playerHealth--
		isPlayerInvincible = true
		invincibilityTimer = invincibilityDuration
		if playerHealth <= 0 {
			gameState = stateGameOver
		}
	}
}

func checkKeyCollection() {
	for i := range keys {
		k := &keys[i]
		if !k.found && int(playerX)+8 > int(k.x) && int(playerX) < int(k.x)+8 && int(playerY)+8 > int(k.y) && int(playerY) < int(k.y)+8 {
			k.found = true
			keysCollected++
			message = "Key collected!"
			messageTimer = messageDisplayTime
		}
	}
}

func checkPoisonItemCollection() {
	for i := range poisonItems {
		f := &poisonItems[i]
		if !f.found && int(playerX)+8 > int(f.x) && int(playerX) < int(f.x)+8 && int(playerY)+8 > int(f.y) && int(playerY) < int(f.y)+8 {
			f.found = true
			hasPoisonPower = true
			message = "Poison orb collected!"
			messageTimer = messageDisplayTime
		}
	}
}
