return {
	{
		"tiagovla/tokyodark.nvim",
		enabled = true,
		opts = {},
		config = function(_, opts)
			require("tokyodark").setup(opts)
			vim.cmd.colorscheme("tokyodark")
		end,
	},
	{
		"rebelot/kanagawa.nvim",
		enabled = true,
		opts = {},
		config = function(_, opts)
			require("kanagawa").setup(opts)
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
