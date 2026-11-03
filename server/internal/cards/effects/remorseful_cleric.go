package effects

// Remorseful Cleric — Creature — Spirit Cleric {1}{W}, 2/1 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Flying
//	 Sacrifice this creature: Exile target player's graveyard."
//
// Angel of Finality's body (exileTargetPlayersGraveyard) behind a
// sacrifice-this cost. The target is a player, chosen as the ability goes
// on the stack, so the graveyard exiled is the one that exists as it
// resolves; the sacrifice is a cost, so the Cleric is gone before anyone
// can respond.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c40bc9f6-2b72-4ae1-b912-50075af59628",
		Name:            "Remorseful Cleric",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Exile target player's graveyard.",
			Cost:    SacrificeThis(),
			Targets: TargetPlayer("target player"),
			Effect:  exileTargetPlayersGraveyard,
		}},
	})
}
