package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b213_helpers.go — shared pieces of Dominaria batch 213 (#5530).

// S82OldB213CastHistoricSpellThisTurn is "you've cast a historic spell this
// turn" (Relic Runner; CR 700.6): one of this turn's cast events is `player`'s
// and the spell it cast is an artifact, a legendary or a Saga. The turn's cast
// log is read, so a historic spell that was countered or has resolved still
// counts, and one cast on another turn does not.
func S82OldB213CastHistoricSpellThisTurn(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCast || ev.Actor != player {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && B09IsHistoric(c) {
			return true
		}
	}
	return false
}

// S82OldB213PowerstoneShards is "the number of artifacts you control named
// Powerstone Shard" for `player`, read through the layers.
func S82OldB213PowerstoneShards(g *game.Game, player uuid.UUID) int {
	return CountControlled(g, player, func(c game.Card) bool {
		return c.IsArtifact() && c.Name == "Powerstone Shard"
	})
}

// S82OldB213TwoHeadedGiantFlips is Two-Headed Giant's attack trigger: "flip two
// coins. If both coins come up heads, this creature gains double strike until
// end of turn. If both coins come up tails, this creature gains menace until
// end of turn." Two uncalled coins (CR 705.2: only the face matters), and the
// gain goes to the creature only if it is still the permanent that attacked.
func S82OldB213TwoHeadedGiantFlips(g *game.Game, item *game.StackItem) error {
	faces, err := g.FlipCoinsForEffect(game.RandomDraw{Player: item.Controller, Source: item.SourceCardID}, 2)
	if err != nil {
		return err
	}
	keyword := ""
	switch {
	case faces[0] && faces[1]:
		keyword = "double strike"
	case !faces[0] && !faces[1]:
		keyword = "menace"
	default:
		return nil
	}
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{keyword},
		Label:    "Two-Headed Giant — gains " + keyword + " until end of turn",
	}.Apply(NewContext(g, item))
}

// S82OldB213RadhaMana is Grand Warlord Radha's trigger body: one mana of {R}
// or {G} (each pick the player's) for each creature its controller controls
// that is attacking, counted as the ability resolves, and kept through the
// turn's steps and phases.
func S82OldB213RadhaMana(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.AttackingTarget != uuid.Nil {
			n++
		}
	}
	if n <= 0 {
		return nil
	}
	return AddMana{
		Amount: ManaN(n, "R", "G"),
		Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
	}.Apply(NewContext(g, item))
}
