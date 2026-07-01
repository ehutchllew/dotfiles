local themes = {
	farout = {
		opts = {
			indent = { char = "▏" },
			scope = { enabled = true },
		},
		create_highlights = function()
			vim.api.nvim_set_hl(0, "Grayscale0", { fg = "#BF472C" })
		end,
		post_setup = function()
			local hooks = require("ibl.hooks")
			hooks.register(hooks.type.SCOPE_HIGHLIGHT, hooks.builtin.scope_highlight_from_extmark)
		end,
		apply_highlights = function(opts)
			opts.scope.highlight = { "Grayscale0" }
		end,
	},
	kanagawa = {
		opts = {
			indent = { char = "" },
			whitespace = { remove_blankline_trail = false },
		},
		create_highlights = function()
			vim.api.nvim_set_hl(0, "Grayscale0", { background = "#1F1F28" })
			vim.api.nvim_set_hl(0, "Grayscale1", { background = "#1C1C25" })
			vim.api.nvim_set_hl(0, "Grayscale2", { background = "#191922" })
			vim.api.nvim_set_hl(0, "Grayscale3", { background = "#16161F" })
			vim.api.nvim_set_hl(0, "Grayscale4", { background = "#13131C" })
			vim.api.nvim_set_hl(0, "Grayscale5", { background = "#101009" })
			vim.api.nvim_set_hl(0, "Grayscale6", { background = "#0E0E06" })
			vim.api.nvim_set_hl(0, "Grayscale7", { background = "#0B0B03" })
		end,
		apply_highlights = function(opts)
			local highlight = {
				"Grayscale0", "Grayscale1", "Grayscale2", "Grayscale3",
				"Grayscale4", "Grayscale5", "Grayscale6", "Grayscale7",
			}
			opts.indent.highlight = highlight
			opts.whitespace.highlight = highlight
		end,
	},
	onedark = {
		opts = {
			indent = { char = "" },
			whitespace = { remove_blankline_trail = false },
		},
		create_highlights = function()
			vim.api.nvim_set_hl(0, "Grayscale0", { background = "#1a212e" })
			vim.api.nvim_set_hl(0, "Grayscale1", { background = "#1d2431" })
			vim.api.nvim_set_hl(0, "Grayscale2", { background = "#202734" })
			vim.api.nvim_set_hl(0, "Grayscale3", { background = "#232a37" })
			vim.api.nvim_set_hl(0, "Grayscale4", { background = "#262d3a" })
			vim.api.nvim_set_hl(0, "Grayscale5", { background = "#29303d" })
			vim.api.nvim_set_hl(0, "Grayscale6", { background = "#2c3340" })
		end,
		apply_highlights = function(opts)
			local highlight = {
				"Grayscale0", "Grayscale1", "Grayscale2", "Grayscale3",
				"Grayscale4", "Grayscale5", "Grayscale6",
			}
			opts.indent.highlight = highlight
			opts.whitespace.highlight = highlight
		end,
	},
	tokyodark = {
		opts = {
			indent = { char = "┊" },
			whitespace = { remove_blankline_trail = false },
		},
		create_highlights = function()
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
		end,
		apply_highlights = function(opts)
			local highlight = {
				"RainbowRed", "RainbowYellow", "RainbowBlue", "RainbowGreen",
				"RainbowViolet", "RainbowCyan", "RainbowOrange",
				"Grayscale0", "Grayscale1", "Grayscale2", "Grayscale3",
				"Grayscale4", "Grayscale5", "Grayscale6", "Grayscale7",
			}
			opts.indent.highlight = highlight
			opts.whitespace.highlight = highlight
		end,
	},
}

local function apply_ibl_theme(colorscheme)
	local hooks = require("ibl.hooks")
	hooks.clear_all()

	local theme = themes[colorscheme]
	if not theme then
		require("ibl").setup({
			indent = { char = "▏" },
			scope = { enabled = true },
		})
		return
	end

	local opts = vim.deepcopy(theme.opts)

	hooks.register(hooks.type.HIGHLIGHT_SETUP, theme.create_highlights)
	theme.create_highlights()
	theme.apply_highlights(opts)

	require("ibl").setup(opts)

	if theme.post_setup then
		theme.post_setup()
	end
end

return {
	{
		"lukas-reineke/indent-blankline.nvim",
		enabled = true,
		main = "ibl",
		config = function()
			local initial = vim.g.colors_name or "farout"
			apply_ibl_theme(initial)

			vim.api.nvim_create_autocmd("ColorScheme", {
				callback = function(args)
					apply_ibl_theme(args.match)
				end,
			})
		end,
	},
}
