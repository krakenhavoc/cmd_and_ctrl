package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Olivia, Crimson Bride — Legendary Creature — Vampire Noble {4}{B}{R},
// 3/4 (Innistrad: Crimson Vow):
//
//	"Flying, haste
//	 Whenever Olivia attacks, return target creature card from your
//	 graveyard to the battlefield tapped and attacking. It gains "When you
//	 don't control a legendary Vampire, exile this creature.""
//
// The return is Interceptor's tapped-and-attacking entry (the card was
// never declared as an attacker, so nothing that watches attacks sees it),
// asking what it attacks when more than one thing could be attacked. Once
// it has entered it gains, with no duration (CR 611.2a), a state trigger
// (CR 603.8) that exiles it whenever its controller controls no legendary
// Vampire. The grant is that permanent's alone: a zone change makes a new
// object that has not got it (CR 400.7).
//
// No simplification.
func init() {
	trigger := On(game.EventAttack, ThisAttacked,
		"Olivia, Crimson Bride — return target creature card from your graveyard to the battlefield tapped and attacking; it gains an exile trigger",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			ids := LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			return vowB17PutAttackingThen(ctx, ids[0], "Olivia, Crimson Bride", oliviaCrimsonBrideGrantTail)
		})
	trigger.Targets = TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())
	Register(Spec{
		OracleID:        "8922a91e-3d15-4351-8d77-e0d6bc4de82e",
		Name:            "Olivia, Crimson Bride",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered:       []game.TriggeredAbility{trigger},
		Grants: []AbilityGrant{{
			Key: oliviaCrimsonBrideExileGrant,
			Triggered: []game.TriggeredAbility{
				WhenYouControlNo(game.PermanentQuery{Subtypes: []string{"Vampire"}, Supertypes: []string{"legendary"}},
					"Olivia, Crimson Bride — exile this creature", ExileThisThen("", nil)),
			},
			Text: "When you don't control a legendary Vampire, exile this creature.",
		}},
	})
}

const oliviaCrimsonBrideExileGrant = "olivia-crimson-bride/exile"

// oliviaCrimsonBrideGrantTail is "It gains …", for the card that entered.
// An on-disk key: never renamed, never reused.
var oliviaCrimsonBrideGrantTail = CardTail("then/olivia-crimson-bride-grant", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.GrantAbilitiesMod(oliviaCrimsonBrideExileGrant)},
			Duration: game.IndefiniteDuration(),
			Label:    "Olivia, Crimson Bride — gains an exile trigger",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
