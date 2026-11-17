package effects

// Konda, Lord of Eiganjo — Legendary Creature — Human Samurai {5}{W}{W},
// 3/3 (Champions of Kamigawa, #5603):
//
//	"Vigilance, indestructible
//	 Bushido 5 (Whenever this creature blocks or becomes blocked, it gets
//	 +5/+5 until end of turn.)"
//
// Two printed keywords and the shared bushido trigger (CR 702.45a).
//
// No simplification.
func init() {
	chkB1Bushido("efea2342-bb24-472a-acfb-1b9e012af91b", "Konda, Lord of Eiganjo", 5, "vigilance", "indestructible")
}
