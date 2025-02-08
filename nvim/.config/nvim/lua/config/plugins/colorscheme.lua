return {
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
	{
		"thallada/farout.nvim",
		enabled = true,
		opts = {
			style = "night",
		},
		config = function(_, opts)
			require("farout").setup(opts)
			vim.cmd.colorscheme("farout")
		end
	},
	{
		"tiagovla/tokyodark.nvim",
		enabled = true,
		opts = {},
		config = function(_, opts)
			require("tokyodark").setup(opts)
		end,
	},
}
