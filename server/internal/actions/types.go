package actions

// allTypes is every action type Dispatch knows: the closed set of
// cmdctrl_actions_total's type label (ADR 0123 §3), which ws registers
// with the metrics package. TestTypesListsEveryTypeConstant fails when a
// Type constant is declared and not listed here.
var allTypes = []Type{
	TypeDrawCard,
	TypePlayCard,
	TypeMoveCard,
	TypeTap,
	TypeUntap,
	TypeUntapAll,
	TypePassPriority,
	TypePassTurn,
	TypeMulligan,
	TypeShuffleLibrary,
	TypeChangeLife,
	TypeAddCounter,
	TypeSetCommanderDamage,
	TypeSetBattlefieldPosition,
	TypeConcede,
	TypeKeepHand,
	TypeDeclareAttacker,
	TypeDeclareAttackers,
	TypeDeclareBlocker,
	TypeDeclareBlockers,
	TypeFinishBlocks,
	TypeClearCombat,
	TypeAdvanceStep,
	TypeSetMonarch,
	TypeSetInitiative,
	TypeSetGoaded,
	TypeSetPoison,
	TypeSetEnergy,
	TypeSetPromise,
	TypeStartVote,
	TypeCastVote,
	TypeEndVote,
	TypeSetUndoLimit,
	TypeSetTableSettings,
	TypeCastSpell,
	TypeCounterSpell,
	TypeCounterAbility,
	TypeActivateAbility,
	TypeActivateLoyalty,
	TypeAnnounceTrigger,
	TypeMarkDamage,
	TypeAddPlayerCounter,
	TypeDiscardSelection,
	TypeResolveChoice,
	TypeSetMaxHandSize,
	TypeActivateManaAbility,
	TypeSpecialAction,
	TypeSacrificePermanent,
	TypeRollOpening,
	TypeHostRollRemaining,
	TypeChooseStartingPlayer,
	TypeRollTableDie,
}

// Types returns every action type, in declaration order. The slice is
// a copy.
func Types() []Type { return append([]Type(nil), allTypes...) }
