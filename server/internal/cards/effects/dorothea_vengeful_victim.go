package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dorothea, Vengeful Victim // Dorothea's Retribution (#1855, ADR 0107
// §4) — a disturb card whose back face is an Aura.
//
// Front face, Legendary Creature — Spirit {W}{U}, 4/4:
//
//	"Flying
//	 When Dorothea attacks or blocks, sacrifice it at end of combat.
//	 Disturb {1}{W}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature attacks, create a
//	 4/4 white Spirit creature token with flying that's tapped and
//	 attacking. Sacrifice that token at end of combat."
//	 If Dorothea's Retribution would be put into a graveyard from
//	 anywhere, exile it instead."
//
// "Attacks or blocks" is one trigger with two conditions, and it
// triggers once however many creatures she blocks (CR 509.3a). Its
// resolution schedules a delayed trigger at the beginning of this
// combat's end of combat step (CR 511.2) that sacrifices Dorothea —
// the same object (CR 400.7): a Dorothea flickered in the meantime is
// not sacrificed.
//
// The Aura grants its quoted ability as a bundle (ADR 0093). The token
// is put onto the battlefield attacking, never declared (CR 508.4), and
// a delayed trigger sacrifices it at end of combat.
//
// Sandbox simplification, declared on the back face: CR 508.4 lets the
// token's controller choose what it attacks; here it attacks the player
// the enchanted creature was declared against — the player behind a
// planeswalker or battle it attacked — as Leonin Warleader's Cats do.
func init() {
	Register(Spec{
		OracleID:         dorotheaOracleID,
		Name:             "Dorothea, Vengeful Victim",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{1}{W}{U}")},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
			}, "Dorothea, Vengeful Victim — sacrifice it at end of combat",
				sacrificeThisAtEndOfCombat("Dorothea, Vengeful Victim — sacrifice it")),
		},
	})
	Register(Spec{
		OracleID:     dorotheaOracleID + "#1",
		Name:         "Dorothea's Retribution",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Spirit token attacks the player the enchanted creature attacked rather than a player of your choice."},
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key: dorotheasRetributionGrant,
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Key: dorotheasRetributionLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, dorotheasRetributionLabel)
					item.Params.Player = b17DefendingPlayer(g, ev)
					return item
				},
				Effect: dorotheasRetributionSpirit,
			}},
			Text: "Whenever this creature attacks, create a 4/4 white Spirit creature token with flying that's tapped and attacking. Sacrifice that token at end of combat.",
		}},
		Static:       []game.StaticAbility{GrantAbilitiesToAttached(dorotheasRetributionGrant)},
		Replacements: []game.ReplacementEffect{DisturbedExile("Dorothea's Retribution")},
	})
}

const (
	dorotheaOracleID          = "2cef4171-8151-4ee9-83a7-bcb5116451bf"
	dorotheasRetributionGrant = "dorotheas-retribution/spirit"
	dorotheasRetributionLabel = "Dorothea's Retribution — a 4/4 Spirit, tapped and attacking"
)
