package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lorehold Archivist // Restore Relic — a preparation card (CR 722,
// ADR 0090), Secrets of Strixhaven Commander (#4834):
//
//	Lorehold Archivist {1}{R}{W}, Creature — Dwarf Artificer 3/2:
//	"First strike
//	 At the beginning of your upkeep, if there are three or more artifact
//	 and/or creature cards in your graveyard, this creature becomes
//	 prepared."
//	Restore Relic {2}{R}{W}, Sorcery:
//	"Exile target artifact or creature card from your graveyard. Create a
//	 token that's a copy of it."
//
// The upkeep trigger's "if" is an intervening if (CR 603.4): checked as
// the upkeep begins and again as the ability resolves (InterveningIf). A
// card that is both an artifact and a creature counts once.
//
// Restore Relic is The Scarab God's exile-then-copy (ExileTargetThen
// TokenCopy) with no exception: only a card that really was exiled is
// copied, and the token copies the card as printed (the ruling).
//
// No simplification.
const loreholdArchivistOracleID = "a371d035-488e-4b3d-81de-ee7859452c0e"

func init() {
	Register(Spec{
		OracleID:        loreholdArchivistOracleID,
		Name:            "Lorehold Archivist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{Name: "three or more artifact and/or creature cards in your graveyard",
				Test: func(ctx *Context) bool { return loreholdArchivistGraveyardCount(ctx.Game, ctx.Controller()) >= 3 }},
				On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return ByYou(ev, source, lki, g) && loreholdArchivistGraveyardCount(g, source.Controller) >= 3
				}, "Lorehold Archivist — becomes prepared (three or more artifact and/or creature cards in your graveyard)",
					socB4ThisBecomesPrepared)),
		},
	})
	Register(Spec{
		OracleID:     loreholdArchivistOracleID + "#1",
		Name:         "Restore Relic",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target artifact or creature card from your graveyard",
			YouOwn(), Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileTargetThenTokenCopy(ctx.Game, item, restoreRelicCopy)
		},
	})
}

// restoreRelicCopy is "create a token that's a copy of it", with no
// exception. An on-disk key.
var restoreRelicCopy = ExileThenTokenCopyTail("exile-then-token-copy/restore-relic", func(*Context) func(t *game.Card) {
	return nil
})

// loreholdArchivistGraveyardCount is the number of artifact and/or
// creature cards in the player's graveyard.
func loreholdArchivistGraveyardCount(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() || c.IsCreature() {
			n++
		}
	}
	return n
}
