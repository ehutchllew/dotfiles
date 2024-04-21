local M = {}

local function config()
	return {
		nmappings = {
			["/"] = { 
				"<Plug>(comment_toggle_linewise_current)",
				"Comment toggle current line"
			},
		},
		nopts = {
			buffer = nil,
			mode = "n",
			noremap = true,
			nowait = true,
			prefix = "<leader>",
			silent = true,
		},
		vmappings = {
			["/"] = { 
				"<Plug>(comment_toggle_linewise_visual)",
				"Comment toggle current line"
			},
		},
		vopts = {
			buffer = nil,
			mode = "v",
			noremap = true,
			nowait = true,
			prefix = "<leader>",
			silent = true,
		},
	}
end

M.setup = function()
	local wk = require("which-key")
		local c = config()
	
	wk.setup()
	wk.register(c.nmappings, c.nopts)
	wk.register(c.vmappings, c.vopts)
end


return M
