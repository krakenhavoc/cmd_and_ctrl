package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b63_helpers.go — GAP batch 63 (#5948): the shared pieces of
// the batch's cards.

// gapB63CreatureYouControlDealtCombatDamageToAPlayer is the When of
// "whenever a creature you control deals combat damage to a player": one
// trigger per creature and per player it hit (CR 510.2, 603.2c), the
// creature judged as it dealt the damage.
func GapB63CreatureYouControlDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return CombatDamageToPlayerBy(ev, source.Controller, g)
}

// gapB63DamageDealer is the creature that dealt the damage the trigger
// fired on, or false if it has left (a creature that dealt damage and then
// died still did).
func GapB63DamageDealer(g *game.Game, item *game.StackItem) (uuid.UUID, bool) {
	if item == nil || item.Trigger == nil || item.Trigger.Event.Source == uuid.Nil {
		return uuid.Nil, false
	}
	id := item.Trigger.Event.Source
	return id, OnBattlefield(g, id)
}

// gapB63OpponentsAttackedThisTurn is how many opponents of `you` they
// attacked with a creature this turn, over every combat of the turn: each
// opponent counts once, an attack on a planeswalker or a battle counts no
// player (Melee's reading, CR 702.121a), and a creature that attacked and has
// since left still counts (the attack records are history).
//
// Caller must hold g.mu.
func GapB63OpponentsAttackedThisTurn(g *game.Game, you uuid.UUID) int {
	seen := map[uuid.UUID]bool{}
	for _, a := range g.TurnTally.Attacks {
		if a.Defender == you || seen[a.Defender] || g.PlayerByIDForEffect(a.Defender) == nil {
			continue
		}
		seen[a.Defender] = true
	}
	return len(seen)
}

// gapB63CountersAmongPlayersAndPermanents is "the number of counters among
// players and permanents": every counter of every kind on every player and
// on every permanent on the battlefield, whoever controls it (Lumbering
// Megasloth). A cost modifier's amount, read as the cost is determined
// (CR 601.2f); the spell itself is on the stack and carries none.
func GapB63CountersAmongPlayersAndPermanents(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	n := 0
	if q.Game.Battlefield != nil {
		for _, c := range q.Game.Battlefield.Cards {
			for _, k := range c.Counters {
				if k > 0 {
					n += k
				}
			}
		}
	}
	for _, p := range q.Game.Seats {
		if p == nil {
			continue
		}
		for _, k := range p.Counters {
			if k > 0 {
				n += k
			}
		}
	}
	return n
}

// gapB63SettlementManaGrant is the bundle key of the mana ability Preston
// Garvey's Settlement gives the land it enchants. On disk.
const gapB63SettlementManaGrant = "settlement/any-color"

// printedGapB63SettlementToken is Preston Garvey's Settlement: a green Aura
// enchantment token with enchant land and "Enchanted land has '{T}: Add one
// mana of any color.'" (Cryptolith Rite's grant, given to the enchanted land
// by a layer 6 static). Its slug is an on-disk identity: never renamed.
func printedGapB63SettlementToken() TokenTemplate {
	return TokenTemplate{
		Slug: "settlement",
		Card: game.Card{
			Name:     "Settlement",
			TypeLine: "Token Enchantment — Aura",
			Colors:   []string{"G"},
		},
		Enchant: TargetPermanent("land", Land()),
		Static:  []game.StaticAbility{GrantAbilitiesToAttached(gapB63SettlementManaGrant)},
		Grants: []AbilityGrant{{
			Key:  gapB63SettlementManaGrant,
			Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}},
			Text: "{T}: Add one mana of any color.",
		}},
		Text: "Enchant land\nEnchanted land has \"{T}: Add one mana of any color.\"",
	}
}

// GapB63SettlementToken is that Settlement's card value for CreateToken.
func GapB63SettlementToken() game.Card { return TokenFromCatalog(printedGapB63SettlementToken) }

// gapB63EachDifferentGraveyard is "from different graveyards": no two picks
// share an owner, so a clause of one target per graveyard (Mysterious
// Stranger) takes one card from each, never two from the same.
func GapB63EachDifferentGraveyard() *game.TargetDifference {
	return &game.TargetDifference{
		Label: "each come from a different graveyard",
		Key: func(_ *game.Game, c game.Card, _ game.ZoneKind) (string, bool) {
			return c.Owner.String(), true
		},
	}
}
