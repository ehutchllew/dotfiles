local M = {}

--[[ local mode_symbols = {
	insert_mode = "i",
	normal_mode = "n",
	term_mode = "t",
	visual_mode = "v",
	visual_block_mode = "x",
	command_mode = "c",
	operator_pending_mode = "o",
} --]]

function M:init()
	require("config.keymaps.global"):init()
	require("config.keymaps.windows"):init()
	require("config.keymaps.buffers"):init()
end

return M
