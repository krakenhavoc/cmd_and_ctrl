package effects

// Desert Twister — Sorcery {4}{G}{G} (Arabian Nights, #5641):
//
//	"Destroy target permanent."
//
// Vindicate's text at green's price: any permanent, lands included.
// Indestructible is honoured by the destroy path.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f880348-6dc8-4cf8-9313-9893c41a70a3",
		Name:         "Desert Twister",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OnResolve:    destroyTheTargetPermanent,
	})
}
