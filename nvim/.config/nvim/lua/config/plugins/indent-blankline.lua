local FarOut = {
	opts = {
		indent = {
			char = "▏",
		},
		scope = { enabled = true },
	},
	config = function(_, opts)
		local highlight = {
			"Grayscale0"
		}

		local hooks = require("ibl.hooks")
		hooks.register(hooks.type.HIGHLIGHT_SETUP, function()
			vim.api.nvim_set_hl(0, "Grayscale0", { fg = "#BF472C" })
		end)

		opts.scope.highlight = highlight
		require("ibl").setup(opts)
		-- NOTE: Remove this next line if youdon't want scope based highlighting
		hooks.register(hooks.type.SCOPE_HIGHLIGHT, hooks.builtin.scope_highlight_from_extmark)
	end
}

local Kanagawa = {
	opts = {
		indent = {
			char = "",
		},
		whitespace = {
			remove_blankline_trail = false,
		},
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
			"Grayscale7",
		}

		vim.api.nvim_set_hl(0, "Grayscale0", { background = "#1F1F28" })
		vim.api.nvim_set_hl(0, "Grayscale1", { background = "#1C1C25" })
		vim.api.nvim_set_hl(0, "Grayscale2", { background = "#191922" })
		vim.api.nvim_set_hl(0, "Grayscale3", { background = "#16161F" })
		vim.api.nvim_set_hl(0, "Grayscale4", { background = "#13131C" })
		vim.api.nvim_set_hl(0, "Grayscale5", { background = "#101009" })
		vim.api.nvim_set_hl(0, "Grayscale6", { background = "#0E0E06" })
		vim.api.nvim_set_hl(0, "Grayscale7", { background = "#0B0B03" })

		opts.indent.highlight = highlight
		opts.whitespace.highlight = highlight
		require("ibl").setup(opts)
	end
}

local OneDark = {
	opts = {
		indent = {
			char = "",
		},
		whitespace = {
			remove_blankline_trail = false,
		},
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

		vim.api.nvim_set_hl(0, "Grayscale0", { background = "#1a212e" })
		vim.api.nvim_set_hl(0, "Grayscale1", { background = "#1d2431" })
		vim.api.nvim_set_hl(0, "Grayscale2", { background = "#202734" })
		vim.api.nvim_set_hl(0, "Grayscale3", { background = "#232a37" })
		vim.api.nvim_set_hl(0, "Grayscale4", { background = "#262d3a" })
		vim.api.nvim_set_hl(0, "Grayscale5", { background = "#29303d" })
		vim.api.nvim_set_hl(0, "Grayscale6", { background = "#2c3340" })

		opts.indent.highlight = highlight
		opts.whitespace.highlight = highlight
		require("ibl").setup(opts)
	end
}

local TokyoDark = {
	opts = {
		indent = {
			char = "┊",
		},
		whitespace = {
			remove_blankline_trail = false,
		},
	},
	config = function(_, opts)
		local highlight = {
			"RainbowRed",
			"RainbowYellow",
			"RainbowBlue",
			"RainbowGreen",
			"RainbowViolet",
			"RainbowCyan",
			"RainbowOrange",
			"Grayscale0",
			"Grayscale1",
			"Grayscale2",
			"Grayscale3",
			"Grayscale4",
			"Grayscale5",
			"Grayscale6",
			"Grayscale7",
		}

		local hooks = require("ibl.hooks")
		hooks.register(hooks.type.HIGHLIGHT_SETUP, function()
			vim.api.nvim_set_hl(0, "RainbowRed", { fg = "#E06C75" })
			vim.api.nvim_set_hl(0, "RainbowYellow", { fg = "#E5C07B" })
			vim.api.nvim_set_hl(0, "RainbowBlue", { fg = "#61AFEF" })
			vim.api.nvim_set_hl(0, "RainbowGreen", { fg = "#98C379" })
			vim.api.nvim_set_hl(0, "RainbowViolet", { fg = "#C678DD" })
			vim.api.nvim_set_hl(0, "RainbowCyan", { fg = "#56B6C2" })
			vim.api.nvim_set_hl(0, "RainbowOrange", { fg = "#D19A66" })
			vim.api.nvim_set_hl(0, "Grayscale0", { background = "#11121D" })
			vim.api.nvim_set_hl(0, "Grayscale1", { background = "#0F101B" })
			vim.api.nvim_set_hl(0, "Grayscale2", { background = "#0D0E19" })
			vim.api.nvim_set_hl(0, "Grayscale3", { background = "#0B0C17" })
			vim.api.nvim_set_hl(0, "Grayscale4", { background = "#090A15" })
			vim.api.nvim_set_hl(0, "Grayscale5", { background = "#070813" })
			vim.api.nvim_set_hl(0, "Grayscale6", { background = "#050611" })
			vim.api.nvim_set_hl(0, "Grayscale7", { background = "#03040F" })
		end)

		opts.indent.highlight = highlight
		opts.whitespace.highlight = highlight
		require("ibl").setup(opts)
	end
}

return {
	{
		"lukas-reineke/indent-blankline.nvim",
		enabled = true,
		main = "ibl",
		opts = FarOut.opts,
		config = FarOut.config
	}
}
