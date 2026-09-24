package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eerie Ultimatum — Sorcery {W}{W}{B}{B}{B}{G}{G} (EDHREC rank
// 1534):
//
//	"Return any number of permanent cards with different names from
//	 your graveyard to the battlefield."
//
// The seven-mana mass reanimation. The choice is the caster's, so it
// is asked through the graveyard picker as an "any number" clause
// over permanent cards the caster owns, and the cards come back
// under their owner's control — the caster's — in announce order,
// each entering through the ordinary reanimation path so its own
// enters-tapped clause and every ETB trigger fire.
//
// "With different names" is the clause's set rule since #1559
// (EachDifferentName): the picker greys a second card of a name
// already picked and the announce gate refuses the pair (CR 601.2c).
// Before that it was enforced as the spell resolved — of two
// same-named picks only the first came back.
//
// One sandbox simplification, declared, weaker than printed: the
// printed text does not target; this clause does, because a
// pick-from-graveyard prompt that is not a target clause does not
// exist here. The difference is that an opponent can exile a chosen
// card in response and that card then stays put — the rest still
// return.
//
// "Any number" includes none, so the spell can be cast with an empty
// graveyard and do nothing.
func init() {
	Register(Spec{
		OracleID:     "5674f6ae-ed5d-441e-a534-b5dd415165fd",
		Name:         "Eerie Ultimatum",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The cards are chosen as targets, so an opponent can exile one in response and it stays in the graveyard.",
		},
		Targets: TargetCardInGraveyard("any number of permanent cards with different names from your graveyard",
			YouOwn(), Permanent()).WithCount(0, 0).EachDifferent(EachDifferentName()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToBattlefield(ctx)
		},
	})
}

// returnLegalGraveyardTargetsToBattlefield puts every still-legal
// graveyard target onto the battlefield under its owner's control, in
// announce order — the body of the set-rule reanimations (#1559):
// Eerie Ultimatum, Agadeem's Awakening, Behold the Sinister Six!. The
// set rule itself is the clause's; by the time this runs the CR 608.2b
// re-check has already dropped any pick that breaks it.
func returnLegalGraveyardTargetsToBattlefield(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
