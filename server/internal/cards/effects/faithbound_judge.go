package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faithbound Judge // Sinner's Judgment (#1855, ADR 0107 §4) — a
// disturb card whose back face is a Curse.
//
// Front face, Creature — Spirit Soldier {1}{W}{W}, 4/4:
//
//	"Defender, flying, vigilance
//	 At the beginning of your upkeep, if this creature has two or fewer
//	 judgment counters on it, put a judgment counter on it.
//	 As long as this creature has three or more judgment counters on
//	 it, it can attack as though it didn't have defender.
//	 Disturb {5}{W}{W}"
//
// Back face, Enchantment — Aura Curse:
//
//	"Enchant player
//	 At the beginning of your upkeep, put a judgment counter on this
//	 Aura. Then if there are three or more judgment counters on it,
//	 enchanted player loses the game.
//	 If Sinner's Judgment would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The upkeep "if" is an intervening if (CR 603.4): asked as the upkeep
// begins and again on resolution. Defender is the only thing that keeps
// a creature from attacking, so "can attack as though it didn't have
// defender" is the keyword absent while the third counter is on it —
// Stalked Researcher's reading. The one place that differs from the
// printing is a card that asks whether the Judge has defender, and none
// in the catalog does.
//
// Defender is therefore not in PrintedKeywords: that list becomes a
// layer-6 grant applied after the card's own statics, which would put
// the keyword back. faithboundJudgeDefender owns it instead — present
// below three counters, gone at three — whether the card was imported
// with Scryfall's keyword list (the baseline) or not.
//
// Disturbed, the Curse targets a player as it is cast (CR 303.4a,
// EnchantPlayer). Its counter goes on through the CR 614 counter
// window, and the loss is read after it lands (CR 104.3e, LoseTheGame):
// three upkeeps after it enters, the enchanted player loses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         faithboundJudgeOracleID,
		Name:             "Faithbound Judge",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "vigilance"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{5}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.Counters[judgmentCounter] <= 2
			}), "Faithbound Judge — put a judgment counter on it", faithboundJudgeUpkeep),
		},
		Static: []game.StaticAbility{faithboundJudgeDefender()},
	})
	Register(Spec{
		OracleID:     faithboundJudgeOracleID + "#1",
		Name:         "Sinner's Judgment",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Sinner's Judgment — put a judgment counter on it; at three, enchanted player loses the game", sinnersJudgmentUpkeep),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Sinner's Judgment")},
	})
}

const (
	faithboundJudgeOracleID = "830e3e37-a80c-4b0e-b9af-393ad4ca01d7"
	judgmentCounter         = "judgment"
)

// faithboundJudgeDefender is the Judge's printed defender together
// with "as long as this creature has three or more judgment counters on
// it, it can attack as though it didn't have defender": a self-only
// layer-6 static that has the keyword below three counters and strips
// it (RemoveFromAttached's removal, aimed at the Judge) at three.
func faithboundJudgeDefender() game.StaticAbility {
	strip := RemoveFromAttached("defender").Apply
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
			if source.Counters[judgmentCounter] >= 3 {
				strip(c, target, g, source)
				return
			}
			appendKeywordsTo(c, []string{"defender"})
		},
	}
}

// faithboundJudgeUpkeep is the upkeep trigger's resolution: the
// intervening if asked again (CR 603.4), then the counter.
func faithboundJudgeUpkeep(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) || b39CountersOn(g, item.SourceCardID, judgmentCounter) > 2 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: judgmentCounter, N: 1}.Apply(NewContext(g, item))
}

// sinnersJudgmentUpkeep puts the counter on the Curse and, once it has
// landed, makes the enchanted player lose with three or more.
func sinnersJudgmentUpkeep(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	curse := item.SourceCardID
	return g.AddCounterThenForEffect(curse, judgmentCounter, 1, func(g *game.Game, _ int) error {
		if b39CountersOn(g, curse, judgmentCounter) < 3 {
			return nil
		}
		c, ok := g.LookupCardForEffect(curse)
		if !ok || c.AttachedTo.Kind != game.TargetPlayer {
			return nil
		}
		return LoseTheGame{Player: c.AttachedTo.ID}.Apply(NewContext(g, item))
	})
}
