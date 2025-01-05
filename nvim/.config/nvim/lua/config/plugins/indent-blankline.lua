local highlight = {
	"CursorColumn",
	"Whitespace",
}
return {
	{
		"lukas-reineke/indent-blankline.nvim",
		enabled = true,
		main = "ibl",
		opts = {
			indent = {
				highlight = highlight,
				char = "",
			},
			whitespace = {
				highlight = highlight,
				remove_blankline_trail = false,
			},
			scope = { enabled = true },
		},
		config = function(_, opts)
			require("ibl").setup(opts)
		end
	}
}
