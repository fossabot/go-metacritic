package metacritic

type Platform string

const (
	// Mobile
	IOS Platform = "9" // Apple iOS

	// Sega
	DC Platform = "15" // Dreamcast

	// Sony
	PS     Platform = "10"    // Playstation
	PS2    Platform = "6"     // Playstation 2
	PS3    Platform = "1"     // Playstation 3
	PS4    Platform = "72496" // Playstation 4
	PSP    Platform = "7"     // Playstation Portable
	PSVita Platform = "67365" // Playstation Vita

	// Nintendo
	GC     Platform = "13"     // GameCube
	GBA    Platform = "11"     // Gameboy Advanced
	N64    Platform = "14"     // Nintendo 64
	N3DS   Platform = "16"     // Nintendo 3DS
	NDS    Platform = "4"      // Nintendo DS
	Switch Platform = "268409" // Nintendo Switch
	Wii    Platform = "8"      // Nintendo Wii
	WiiU   Platform = "68410"  // Nintendo Wii U

	// Microsoft
	PC      Platform = "3"     // Personal Computer
	Xbox    Platform = "12"    // Xbox
	Xbox360 Platform = "2"     // Xbox 360
	XboxOne Platform = "80000" // Xbox One
)
