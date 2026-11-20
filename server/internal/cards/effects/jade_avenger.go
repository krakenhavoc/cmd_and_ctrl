package effects

// Jade Avenger — Creature — Frog Samurai {1}{G}, 2/2 (Modern Horizons 2,
// #5501):
//
//	"Bushido 2 (Whenever this creature blocks or becomes blocked, it gets
//	 +2/+2 until end of turn.)"
//
// Samurai Enforcers' bushido 2 (chkB1Bushido, CR 702.45a).
//
// No simplification.
func init() {
	chkB1Bushido("177248b7-f3fd-42f8-b325-2625a2bb28ff", "Jade Avenger", 2)
}
