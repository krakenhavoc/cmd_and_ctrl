package effects

// otj_r_helpers.go — the shared pieces behind the Outlaws of Thunder
// Junction R cards (#4468). Every package-level name carries the otjR
// prefix because other batches land beside this one.

// otjRDrawThenTreasure is "Draw a card and create a Treasure token" in
// the printed order (Reckless Lackey, Redrock Sentinel): the Treasure
// waits for any prompt the draw asks (CR 608.2c, #4155).
func otjRDrawThenTreasure() Effect {
	return Do(DrawCards{N: 1}, CreateToken{Template: TreasureToken(), N: 1})
}
