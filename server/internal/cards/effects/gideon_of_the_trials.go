package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gideon of the Trials — Legendary Planeswalker — Gideon {1}{W}{W},
// starting loyalty 3:
//
//	"+1: Until your next turn, prevent all damage target permanent
//	 would deal.
//	 0: Until end of turn, Gideon becomes a 4/4 Human Soldier creature
//	 with indestructible that's still a planeswalker. Prevent all damage
//	 that would be dealt to him this turn.
//	 0: You get an emblem with "As long as you control a Gideon
//	 planeswalker, you can't lose the game and your opponents can't win
//	 the game.""
//
// Three abilities, three shapes that already exist:
//
//   - THE +1 is ADR 0108 §7's source shield (`preventFromSource`) with
//     the targeted permanent as its source, pinned at resolution with no
//     prompt, protecting everything, lasting until Gideon's controller's
//     next turn (DamageShield.UntilYourNextTurn, CR 611.2b). A target
//     that has left the battlefield is illegal (CR 608.2b) and nothing is
//     prevented.
//   - THE FIRST 0 is one data record pinned to Gideon until end of turn
//     (Creature, Human Soldier, base 4/4, indestructible; he keeps his
//     planeswalker type), then a source shield with no source protecting
//     Gideon himself this turn, so damage to him removes no loyalty and
//     marks no damage. A Gideon that left the battlefield in response is
//     a new object (CR 400.7) and nothing happens.
//   - THE EMBLEM is ADR 0109 §5's EmblemSpec.GameEndGates: Platinum
//     Angel's two gates, each with a While read off the emblem — its
//     owner (CR 114.2) controls a planeswalker with the subtype Gideon.
//     The gate is read where the loss or win would happen (CR 104.3,
//     614.17), so the moment the last Gideon leaves, a player at 0 life
//     loses at the next check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9bbaad7-c016-4908-a6a7-2c26112a6bf6",
		Name:         "Gideon of the Trials",
		Completeness: CompletenessFull,
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label:        "Gideon of the Trials emblem",
			Text:         "As long as you control a Gideon planeswalker, you can't lose the game and your opponents can't win the game.",
			GameEndGates: gideonOfTheTrialsGates(),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Until your next turn, prevent all damage target permanent would deal.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("target permanent"),
				Effect:  gideonOfTheTrialsPlusOne,
			},
			{
				Label:  "0: Until end of turn, Gideon becomes a 4/4 Human Soldier creature with indestructible that's still a planeswalker. Prevent all damage that would be dealt to him this turn.",
				Cost:   LoyaltyCost(0),
				Effect: gideonOfTheTrialsAnimate,
			},
			{
				Label: "0: You get an emblem with \"As long as you control a Gideon planeswalker, you can't lose the game and your opponents can't win the game.\"",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// gideonOfTheTrialsPlusOne is the +1: a shield against all damage the
// target would deal, until its controller's next turn.
func gideonOfTheTrialsPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return PreventDamageFromSource{
			From:              t.ID,
			Protect:           ShieldAnything,
			UntilYourNextTurn: true,
			Label:             "Gideon of the Trials — prevent all damage it would deal until your next turn",
		}.Apply(ctx)
	}
	return nil
}

// gideonOfTheTrialsAnimate is the first 0: Gideon becomes a 4/4 Human
// Soldier creature with indestructible until end of turn, and all damage
// that would be dealt to him this turn is prevented.
func gideonOfTheTrialsAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if z := g.FindCardZoneForEffect(self); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := (ScopedEffectFor{
		Target: self,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod("Human", "Soldier"),
			game.SetBasePowerMod(4),
			game.SetBaseToughnessMod(4),
			game.AddKeywordsMod("indestructible"),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Gideon of the Trials — a 4/4 Human Soldier creature with indestructible until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
		EffectSource:     self,
		Controller:       item.Controller,
		ProtectPermanent: self,
		Label:            "Gideon of the Trials — prevent all damage that would be dealt to him this turn",
	})
	return nil
}

// gideonOfTheTrialsGates is the emblem's "As long as you control a
// Gideon planeswalker, you can't lose the game and your opponents can't
// win the game": Platinum Angel's two gates, each held by the emblem
// owner controlling a Gideon planeswalker.
func gideonOfTheTrialsGates() []game.GameEndGate {
	gates := YouCantLoseOpponentsCantWin()
	for i := range gates {
		gates[i].While = youControlAGideonPlaneswalker
	}
	return gates
}

// youControlAGideonPlaneswalker reports whether the source's controller
// controls a planeswalker with the subtype Gideon, read off the
// effective characteristics.
func youControlAGideonPlaneswalker(g *game.Game, source game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == source.Controller && c.IsPlaneswalker() && c.HasSubtype("Gideon") {
			return true
		}
	}
	return false
}
