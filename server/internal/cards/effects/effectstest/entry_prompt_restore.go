package effectstest

// entry_prompt_restore_test.go — #4253 slice 2: the entry questions that
// pause through the CR 616 resume frame, and the counter, life and
// keyword-action events with no closure tail, are restore points. Each
// test takes a STRICT restore (RestoreStrict, the boot path) while the
// prompt is open, answers it on the restored table, and checks what a
// player sees afterwards. The frame itself is game/replacement_resume_keyed.go.

const RhoxFaithmenderOracle = "2abe9303-d498-4aad-b6b2-8b5064bd2ffd"
