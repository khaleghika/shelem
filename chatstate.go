package main

type GameState int

const (
	StateStart GameState = iota
	StateSelectTrump
	StateSelectHand
	StateInputOtherScore
)

type Team int

const (
	RedTeam Team = iota
	BlackTeam
)

type GameItem struct {
	TrumpTeam     Team
	Claim         int
	OpponentScore int
	TrumpScore    int
}

type Game struct {
	State          GameState
	Items          []*GameItem
	RedTeamScore   int
	BalckTeamScore int
}
