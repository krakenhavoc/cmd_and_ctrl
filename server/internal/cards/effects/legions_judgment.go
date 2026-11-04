package effects

// Legion's Judgment — Sorcery {2}{W} (Jumpstart Azorius deck request,
// #4894):
//
//	"Destroy target creature with power 4 or greater."
//
// The power is the creature's effective power, judged at announce and
// again at resolution (CR 608.2b): shrunk below 4 in response, it is no
// longer a legal target and the spell does nothing. Regeneration and
// indestructible apply as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dce4d65f-cb91-48a2-a2a7-013e73dfb667",
		Name:         "Legion's Judgment",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with power 4 or greater", PowerGE(4)),
		OnResolve:    destroyTheTargetPermanent,
	})
}
