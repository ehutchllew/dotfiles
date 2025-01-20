return {
	{
		"echasnovski/mini.nvim",
		enabled = true,
		opts = {
			hipatterns = {
				highlighters = {
					-- Highlight Standalone 'FIXME', 'HACK', 'TODO', 'NOTE'
					fixme = { pattern = "%f[%w]()FIXME()%f[%W]", group = "MiniHipatternsFixme" },
					hack = { pattern = "%f[%w]()HACK()%f[%W]", group = "MiniHipatternsHack" },
					todo = { pattern = "%f[%w]()TODO()%f[%W]", group = "MiniHipatternsTodo" },
					note = { pattern = "%f[%w]()NOTE()%f[%W]", group = "MiniHipatternsNote" },
				}
			},
			statusline = {
				use_icons = true,
				tabpage_section = "right"
			}
		},
		config = function(_, opts)
			-- Highlight Patterns
			local hipatterns = require("mini.hipatterns")
			local extHghlght = vim.tbl_extend("keep", opts.hipatterns.highlighters, {
				-- Highlight Hex Color Strings (`#rrggbb`)
				hex_color = hipatterns.gen_highlighter.hex_color()
			})
			opts.hipatterns.highlighters = extHghlght
			hipatterns.setup(opts.hipatterns)

			-- Auto Pairs
			local pairs = require("mini.pairs")
			pairs.setup()

			-- Split Join
			local splitjoin = require("mini.splitjoin")
			splitjoin.setup()

			-- Status Line
			local statusline = require("mini.statusline")
			statusline.setup(opts.statusline)

			-- Surround Utility
			local surround = require("mini.surround")
			surround.setup()
		end
	}
}
