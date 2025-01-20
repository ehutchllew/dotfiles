return {
	{
		"echasnovski/mini.nvim",
		enabled = true,
		opts = {
			statusline = {
				use_icons = true,
				tabpage_section = "right"
			}
		},
		config = function(_, opts)
			local pairs = require("mini.pairs")
			pairs.setup()
			local splitjoin = require("mini.splitjoin")
			splitjoin.setup()
			local statusline = require("mini.statusline")
			statusline.setup(opts.statusline)
			local surround = require("mini.surround")
			surround.setup()
			--local tabline = require("mini.tabline")
			--tabline.setup(opts)
		end
	}
}
