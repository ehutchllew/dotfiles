return {
	{
		"rebelot/kanagawa.nvim",
		enabled = true,
		opts = {},
		config = function(_, opts)
			require("kanagawa").setup(opts)
			vim.cmd.colorscheme("kanagawa")
		end
	},
	{
		"navarasu/onedark.nvim",
		enabled = true,
		opts = {
			style = "deep",
		},
		config = function(_, opts)
			require("onedark").setup(opts)
		end
	},
}
