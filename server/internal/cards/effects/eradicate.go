package effects

// Eradicate — Sorcery {2}{B}{B}:
//
//	"Exile target nonblack creature. Search its controller's graveyard,
//	 hand, and library for all cards with the same name as that creature
//	 and exile them. Then that player shuffles."
//
// The creature's name and controller are read as it last existed on the
// battlefield, before it is exiled, so a copy is searched for by the name
// it copied and a stolen creature sends the caster through its
// controller's zones, not its owner's. The caster searches in one prompt,
// seeing the whole hand as they do: every card with that name in the
// graveyard is exiled, a public zone having no fail to find (CR 701.23b),
// and the caster may leave matches in the hand and the library, which are
// hidden. The library is shuffled whatever was taken
// (exilePermanentThenExtract, #4033). The search happens even if the
// creature did not reach exile; a target that is no longer legal makes the
// spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "782ff740-ff43-4c07-af0b-433cb9770661",
		Name:         "Eradicate",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve:    exilePermanentThenExtract("Eradicate"),
	})
}
