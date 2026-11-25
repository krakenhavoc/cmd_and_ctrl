package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akroma, Vision of Ixidor — {5}{W}{W} Legendary Creature — Angel 6/6
// (Commander Legends, S82 OLD b076, #6162):
//
//	"Flying, first strike, vigilance, trample
//	 At the beginning of each combat, until end of turn, each other
//	 creature you control gets +1/+1 if it has flying, +1/+1 if it has
//	 first strike, and so on for double strike, deathtouch, haste,
//	 hexproof, indestructible, lifelink, menace, protection, reach,
//	 trample, vigilance, and partner.
//	 Partner"
//
// The bonus is counted for each creature as the ability resolves, from
// its layered abilities, and does not change for the rest of the turn if
// it gains or loses one. Each listed ability counts once however many
// instances it has; any "hexproof from [quality]" is hexproof (CR
// 702.11d), any protection is protection, and partner is #6162's keyword:
// partner, partner—[text] and partner with [name] (CR 702.124n). Partner
// on Akroma itself is deck construction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b9be5ce9-08bb-4361-940c-9031e429ba86",
		Name:            "Akroma, Vision of Ixidor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "vigilance", "trample", game.KeywordPartner},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat("Akroma, Vision of Ixidor — each other creature you control gets +1/+1 for each listed ability it has",
				effects.Do(effects.Instruction(akromaVisionBoost))),
		},
	})
}

// akromaVisionKeywords are the abilities Akroma counts that are a plain
// keyword token; hexproof and protection are read with their variants.
var akromaVisionKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "haste",
	"indestructible", "lifelink", "menace", "reach", "trample", "vigilance",
	game.KeywordPartner,
}

// akromaVisionCount is how many of Akroma's listed abilities `c` has.
func akromaVisionCount(c *game.Card) int {
	n := 0
	for _, kw := range akromaVisionKeywords {
		if game.HasKeyword(c, kw) {
			n++
		}
	}
	if game.HasKeyword(c, "hexproof") || game.HasHexproofFrom(c) {
		n++
	}
	if game.HasProtection(c) {
		n++
	}
	return n
}

func akromaVisionBoost(ctx *effects.Context) error {
	yours := effects.And(effects.Creature(), effects.YouControl())
	for _, c := range effects.MatchingBattlefield(ctx, yours) {
		if c.InstanceID == ctx.Source() {
			continue
		}
		n := akromaVisionCount(&c)
		if n == 0 {
			continue
		}
		if err := (effects.BoostUntilEOT{
			Target:    c.InstanceID,
			Power:     n,
			Toughness: n,
			Label:     "Akroma, Vision of Ixidor — +1/+1 for each listed ability",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
