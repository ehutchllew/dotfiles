return {
	{
		"lukas-reineke/indent-blankline.nvim",
		enabled = true,
		main = "ibl",
		opts = {
			indent = {
				char = "",
			},
			whitespace = {
				remove_blankline_trail = false,
			},
			scope = { enabled = true },
		},
		config = function(_, opts)
			local highlight = {
				"Grayscale0",
				"Grayscale1",
				"Grayscale2",
				"Grayscale3",
				"Grayscale4",
				"Grayscale5",
				"Grayscale6",
			}

			local hooks = require("ibl.hooks")
			hooks.register(hooks.type.HIGHLIGHT_SETUP, function()
				-- NOTE: When changing colorscheme, update these to maintain a gradient-like indent effect.
				-- KANAGAWA
				vim.api.nvim_set_hl(0, "Grayscale0", { background = "#16161D" })
				vim.api.nvim_set_hl(0, "Grayscale1", { background = "#1A1A21" })
				vim.api.nvim_set_hl(0, "Grayscale2", { background = "#1E1E25" })
				vim.api.nvim_set_hl(0, "Grayscale3", { background = "#222229" })
				vim.api.nvim_set_hl(0, "Grayscale4", { background = "#26262D" })
				vim.api.nvim_set_hl(0, "Grayscale5", { background = "#2A2A31" })
				vim.api.nvim_set_hl(0, "Grayscale6", { background = "#2E2E35" })

				-- ONEDARK
				-- vim.api.nvim_set_hl(0, "Grayscale0", { background = "#1a212e" })
				-- vim.api.nvim_set_hl(0, "Grayscale1", { background = "#1d2431" })
				-- vim.api.nvim_set_hl(0, "Grayscale2", { background = "#202734" })
				-- vim.api.nvim_set_hl(0, "Grayscale3", { background = "#232a37" })
				-- vim.api.nvim_set_hl(0, "Grayscale4", { background = "#262d3a" })
				-- vim.api.nvim_set_hl(0, "Grayscale5", { background = "#29303d" })
				-- vim.api.nvim_set_hl(0, "Grayscale6", { background = "#2c3340" })
			end)

			opts.indent.highlight = highlight
			opts.whitespace.highlight = highlight
			require("ibl").setup(opts)
		end
	}
}
