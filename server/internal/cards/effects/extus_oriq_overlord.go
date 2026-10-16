package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extus, Oriq Overlord // Awaken the Blood Avatar — a modal double-faced
// card (oracle 0b299983…), one file for both faces.
//
// Front, Legendary Creature — Human Warlock {1}{W}{B}{B}, 2/4:
//
//	"Double strike
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 return target nonlegendary creature card from your graveyard to
//	 your hand."
//
// Back, Sorcery {6}{B}{R} (key "<oracle>#1", the MDFC face keyspace):
//
//	"As an additional cost to cast this spell, you may sacrifice any
//	 number of creatures. This spell costs {2} less to cast for each
//	 creature sacrificed this way.
//	 Each opponent sacrifices a creature of their choice. Create a 3/6
//	 black and red Avatar creature token with haste and "Whenever this
//	 token attacks, it deals 3 damage to each opponent.""
//
// The magecraft is the shared Magecraft trigger (#2090): a copy of an
// instant or sorcery you make triggers it as well as a cast, once per
// copy. The Strixhaven ruling that a copy of a card cast from another
// zone triggers it once, as a cast, is the trigger's own rule.
//
// The back face is Torgaar's cost (ADR 0100 §3): the count is announced
// at CR 601.2b and CostsLessPerSacrificed reads it at 601.2f, so the
// discount reaches generic mana only and a fourth creature buys nothing.
// The edict runs first, and the Avatar is made once every opponent has
// chosen and sacrificed, in the order the sentence prints. The Avatar's
// attack trigger lives on the token (ADR 0083), so the damage is dealt
// by the Avatar and keeps working for whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        extusOracleID,
		Name:            "Extus, Oriq Overlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Magecraft("Extus, Oriq Overlord — return target nonlegendary creature card from your graveyard to your hand (magecraft)",
					returnTargetCardToHand),
				TargetCardInGraveyard("target nonlegendary creature card in your graveyard", YouOwn(), Creature(), Not(Legendary())),
			),
		},
	})
	Register(Spec{
		OracleID:       extusOracleID + "#1",
		Name:           "Awaken the Blood Avatar",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("any number of creatures", Creature()),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerSacrificed("{2}", "This spell costs {2} less to cast for each creature sacrificed this way"),
		},
		Purpose: game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				ExceptController: true,
				Match:            Creature(),
				Label:            "a creature",
				Then: func(ctx *Context, _ game.PromptedSacrifices) error {
					return CreateToken{Template: BloodAvatarToken(), N: 1}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}

const extusOracleID = "0b299983-9f0f-404a-acd1-8f142572b1f1"

// BloodAvatarToken is Awaken the Blood Avatar's 3/6 black and red Avatar
// with haste and "Whenever this token attacks, it deals 3 damage to each
// opponent."
func BloodAvatarToken() game.Card { return tokenFromCatalog(printedBloodAvatarToken) }

// printedBloodAvatarToken is that Avatar as PRINTED. The slug is
// qualified because other cards make Avatars that differ from it.
func printedBloodAvatarToken() tokenTemplate {
	return tokenTemplate{
		Slug: "avatar-blood",
		Card: game.Card{
			Name:      "Avatar",
			TypeLine:  "Token Creature — Avatar",
			Power:     3,
			Toughness: 6,
			Colors:    []string{"B", "R"},
			Keywords:  []string{"haste"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Avatar — 3 damage to each opponent", Do(DamageEachOpponent{N: 3})),
		},
		Text: "Haste\nWhenever this token attacks, it deals 3 damage to each opponent.",
	}
}
