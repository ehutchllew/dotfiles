return {
	{
		'stevearc/oil.nvim',
		opts = {
			view_options = {
				show_hidden = true
			}
		},
		-- Optional dependencies
		dependencies = { { "echasnovski/mini.icons", opts = {} } },
		config = function(_, opts)
			local oil = require("oil")
			oil.setup(opts)

			vim.keymap.set("n", "<space>e", function() oil.toggle_float(vim.uv.cwd()) end,
				{ desc = "Toggle Oil Navigator at Root" })
			vim.keymap.set("n", "<space>E", function() oil.toggle_float() end, { desc = "Toggle Oil Navigator at CWD" })
		end
	},
}
