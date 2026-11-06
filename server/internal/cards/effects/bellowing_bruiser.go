package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bellowing Bruiser // Beat a Path — Creature — Ogre {4}{R}, 4/4 //
// Sorcery — Adventure {2}{R} (Wilds of Eldraine, #4471):
//
//	Bellowing Bruiser — "Haste"
//	Beat a Path — "Up to two target creatures can't block this turn.
//	 (Then exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). Beat a Path is Daring Discovery's block restriction over up
// to two targets: the until-end-of-turn "can't block" record goes on each
// target still legal as the spell resolves, and with no target at all the
// Adventure still resolves and goes to exile.
//
// No simplification.
const bellowingBruiserOracleID = "cdb84830-faf2-4beb-8d57-ad696d75876b"

func init() {
	Register(Spec{
		OracleID:        bellowingBruiserOracleID,
		Name:            "Bellowing Bruiser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
	Register(Spec{
		OracleID:     bellowingBruiserOracleID + "#1",
		Name:         "Beat a Path",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Beat a Path — can't block"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
