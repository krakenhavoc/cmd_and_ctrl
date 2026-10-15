package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alchemist's Talent — Enchantment — Class {3}{R}:
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, create two tapped Treasure tokens.
//	 {1}{R}: Level 2
//	 Treasures you control have '{T}, Sacrifice this artifact: Add
//	 two mana of any one color.'
//	 {4}{R}: Level 3
//	 Whenever you cast a spell, if mana from a Treasure was spent to
//	 cast it, this Class deals damage equal to that spell's mana
//	 value to each opponent."
//
// Level 1 and level 2 are ordinary ADR 0071 Class machinery: the
// level-1 ETB and a level-2 grant gated `Level(2)` — which, per CR
// 716.2a, stays active at level 3 too, exactly like the printed
// cumulative levels. The grant's own cost has a component ADR 0093's
// bundle already carries: "Sacrifice this artifact" is
// ManaAbilityCost.Sacrifice, the Treasure sacrificing ITSELF, not the
// Class.
//
// Level 3 (#3172) reads the cast spell's payment record: every mana
// token carries a snapshot of its source's kinds (#1212), so "if mana
// from a Treasure was spent to cast it" is
// StackItemPaidForEffect(spell).Spent().FromTreasure(), checked as an
// intervening if. The damage is the spell's mana value, read when the
// trigger resolves.
func init() {
	const grant = "alchemists-talent/treasure-mana"
	Register(Spec{
		OracleID:       "5a2dfff5-9dba-42f5-bcca-918c16f23807",
		Name:           "Alchemist's Talent",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}, Sacrifice this artifact: Add two mana of any one color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{{
			Layer:          game.Layer6Ability,
			AppliesTo:      treasuresYouControl,
			GrantAbilities: []string{grant},
			ActiveWhen:     Level(2),
		}},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{R}")),
			LevelUp(3, ManaCost("{4}{R}")),
		},
		Triggered: []game.TriggeredAbility{
			AtLevel(3, On(game.EventCast, spellPaidWithTreasureMana,
				"Alchemist's Talent — damage equal to that spell's mana value to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := triggeringSpellManaValue(g, item)
					return g.DamageInstanceForEffect(func() error {
						for _, opp := range ctx.Opponents() {
							if err := (DealDamage{Source: item.SourceCardID, Target: opp, Amount: n}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				})),
			WhenThisEnters("Alchemist's Talent — create two tapped Treasure tokens",
				Do(CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 2})),
		},
	})
}

// spellPaidWithTreasureMana is the level-3 condition (an intervening if,
// CR 603.4): you cast a spell and mana from a Treasure was spent on it.
// The cast announcement's payment record keeps the mana's source kinds as
// a snapshot, so a Treasure that sacrificed itself to pay still counts
// (mana_source.go). A payment the engine did not see reads as "no".
func spellPaidWithTreasureMana(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	return g.StackItemPaidForEffect(ev.CardID).Spent().FromTreasure()
}

// treasuresYouControl is "Treasures you control" — Alchemist's
// Talent's level-2 grant.
func treasuresYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Treasure")
}
