package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Howlsquad Heavy — Creature — Goblin Mercenary {2}{R}, 2/3:
//
//	"Start your engines!
//	 Other Goblins you control have haste.
//	 At the beginning of combat on your turn, create a 1/1 red Goblin
//	 creature token. That token attacks this combat if able.
//	 Max speed — {T}: Add {R} for each Goblin you control."
//
// ADR 0138 (#2122, #2711). The haste is a tribal layer 6 grant over other
// Goblins you control, so the new token has it the moment it enters. The
// attack requirement is pinned to that one token until end of turn, the
// shape Legion Warboss uses. The mana ability counts Goblin creatures as
// it is activated (CR 605.3b) and is available only at max speed.
//
// No simplification.
func init() {
	goblins := TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "cd45a29f-73bc-435c-9d50-09a4ecaa77be",
		Name:            "Howlsquad Heavy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Static:          []game.StaticAbility{TribalKeywordGrant(goblins, "haste")},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Howlsquad Heavy — create a 1/1 red Goblin that attacks this combat if able",
				howlsquadHeavyToken),
		},
		ManaAbilities: []ManaAbility{
			MaxSpeedMana(ManaAbility{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: ProducedPerPermanent("R", MatchCreatureSubtype("Goblin")),
				Label:        "Max speed — {T}: Add {R} for each Goblin you control.",
			}),
		},
	})
}

// howlsquadHeavyToken creates the 1/1 red Goblin and pins "attacks this
// combat if able" to that one token (the haste is the static's).
func howlsquadHeavyToken(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cursor := b25LastEventSeq(g)
	if err := (CreateToken{Template: RedGoblinToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	for _, token := range b27TokensCreatedByAfter(g, item.Controller, cursor) {
		return ScopedEffectFor{
			Target:   token,
			Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Howlsquad Heavy — the token attacks this combat if able",
		}.Apply(ctx)
	}
	return nil
}
