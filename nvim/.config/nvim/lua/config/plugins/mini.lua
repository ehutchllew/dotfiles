return {
	{
		"echasnovski/mini.nvim",
		enabled = true,
		opts = {
			use_icons = true,
			tabpage_setion = "right"
		},
		config = function(_, opts)
			local splitjoin = require("mini.splitjoin")
			splitjoin.setup()
			local statusline = require("mini.statusline")
			statusline.setup(opts)
			local surround = require("mini.surround")
			surround.setup()
			--local tabline = require("mini.tabline")
			--tabline.setup(opts)
		end
	}
}
