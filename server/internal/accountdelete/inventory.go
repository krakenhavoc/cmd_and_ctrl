package accountdelete

// Treatment is what account deletion does to a column.
type Treatment string

const (
	// Deleted: the row is removed (or, for a column of a row that stays,
	// the value is set to NULL).
	Deleted Treatment = "deleted"
	// Anonymised: the row stays and the value is replaced by
	// DeletedPlayerName or by a pointer to the tombstone.
	Anonymised Treatment = "anonymised"
	// Kept: the value stays, because it is part of other people's game
	// record or names no one. Note says which.
	Kept Treatment = "kept"
	// NotPersonal: the column holds nothing about a person.
	NotPersonal Treatment = "not personal"
)

// Entry is one column of the schema and what deletion does to it.
type Entry struct {
	Table     string
	Column    string
	Treatment Treatment
	// Note is why, in a sentence. Required for Kept and Anonymised.
	Note string
}

func cols(table string, t Treatment, note string, names ...string) []Entry {
	out := make([]Entry, 0, len(names))
	for _, n := range names {
		out = append(out, Entry{Table: table, Column: n, Treatment: t, Note: note})
	}
	return out
}

func join(parts ...[]Entry) []Entry {
	var out []Entry
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// Inventory is the schema, column by column, and what deletion does to
// each. It is the maintenance rule of ADR 0149 made executable:
// TestInventoryCoversTheSchema fails when a migration adds a table or
// column missing from it, and TestDeleteTreatsEveryInventoriedColumn
// checks the Deleted and Anonymised entries against a seeded account.
//
// Adding a migration that stores anything about a person? Add its
// columns here with a treatment, make Delete honour it, and update the
// privacy page (client/src/routes/Privacy.svelte) in the same change.
var Inventory = join(
	cols("schema_migrations", NotPersonal, "", "version", "name", "applied_at"),

	cols("games", Kept, "the table's name and clock are part of the record the other players keep",
		"id", "name", "state", "created_at", "started_at", "ended_at", "archived_at", "winner_seat", "outcome", "host_player_id"),
	cols("games", Anonymised, "points at the tombstone, which carries no identity",
		"created_by"),
	cols("games", Deleted, "set to NULL when it was the deleted person's Discord id",
		"host_discord_id"),

	cols("seats", Kept, "the seat keeps its place in the game",
		"game_id", "seat", "player_id", "bot_tier", "agent_client", "deck_name"),
	cols("seats", Anonymised, "the seat's name becomes \"Deleted player\"",
		"guest_name"),
	cols("seats", Anonymised, "points at the tombstone, which carries no identity",
		"user_id"),
	cols("seats", Deleted, "set to NULL when the deck is deleted",
		"deck_id", "pending_discord_id"),

	cols("invites", NotPersonal, "an invite token hash and its clock",
		"token_hash", "game_id", "kind", "created_at", "expires_at", "revoked_at"),
	cols("invites", Anonymised, "points at the tombstone, which carries no identity",
		"created_by"),

	cols("users", Kept, "the tombstone keeps its id and creation time so other tables stay consistent",
		"id", "created_at", "deleted_at"),
	cols("users", Anonymised, "replaced by \"Deleted player\" / NULL / now",
		"display_name", "avatar_url", "last_seen_at", "last_deck", "playmat_id", "playmat_wash", "admin_mode_at",
		"access_checked_at", "access_result"),
	cols("users", Kept, "the revocation watermark must outlive the deletion, or a token issued before it works again after a restart",
		"sessions_invalid_before"),

	cols("identities", Deleted, "the whole row, including the encrypted refresh token",
		"provider", "subject", "user_id", "display_name", "avatar_hash", "refresh_token", "scopes", "linked_at", "refreshed_at"),

	cols("decks", Deleted, "every deck the person saved",
		"id", "owner_id", "name", "source_format", "source_text", "commanders", "card_count", "created_at", "updated_at", "source_url"),

	cols("deck_requests", NotPersonal, "which GitHub issue tracks a deck; no requester",
		"deck_key", "issue_number", "issue_url", "created_at"),
	cols("deck_request_asks", Deleted, "the person's rows, keyed by their Discord id",
		"requester"),
	cols("deck_request_asks", NotPersonal, "which deck was asked about, and when",
		"deck_key", "at"),

	cols("user_settings", Deleted, "the person's synced settings",
		"user_id", "version", "revision", "body", "updated_at"),
	cols("table_setups", Deleted, "the person's last table setup",
		"user_id", "body", "game_id", "updated_at"),
	cols("user_playmats", Deleted, "the rows; the image files are removed by the caller",
		"user_id", "slot", "playmat_id", "width", "height", "created_at"),

	cols("table_blocks", NotPersonal, "which table refused to seat someone again; the game it belongs to",
		"game_id", "created_at", "entry_id"),
	cols("table_blocks", Deleted, "every row of an entry that has a row whose subject is the person's account id or Discord id, so the display-name label and a spectator-session marker under the same entry go with it (a folded guest name names no account and stays)",
		"kind", "subject", "label"),

	cols("player_reports", Deleted, "every report the person filed or is the subject of, whole",
		"id", "reporter_player_id", "reporter_user_id", "reporter_name",
		"reported_player_id", "reported_user_id", "reported_name", "reason"),
	cols("player_reports", Kept, "a recent-chat excerpt carries other players' lines under their seat names; it goes with its report at 90 days, and a report that names neither the deleted person as reporter nor as subject keeps its excerpt as it was",
		"excerpt"),
	cols("player_reports", NotPersonal, "when and where; the table's name is already part of the game record",
		"created_at", "game_id", "game_name"),

	cols("bot_model_budget", NotPersonal, "server-wide token counters",
		"key", "value", "updated_at"),
)
