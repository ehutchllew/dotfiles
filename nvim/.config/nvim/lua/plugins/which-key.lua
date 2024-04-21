local M = {}

local function config()
	return {
		nmappings = {
			["/"] = { 
				"<Plug>(comment_toggle_linewise_current)",
				"Comment toggle current line"
			},
			["h"] = {
				"<cmd>nohlsearch<CR>", "Dismiss Highlight"
			},
			s = {
				-- NOTE: if using `Telescope` is too slow we can change this
				-- to use the telescope.builtin
				name = "Search",
				b = { "<cmd>Telescope current_buffer_fuzzy_find<cr>", "[S]earch Current [B]uffer" },
				f = { "<cmd>Telescope find_files<cr>", "[S]earch [F]iles" },
				h = { "<cmd>Telescope help_tags<cr>", "[S]earch [H]elp" },
				k = { "<cmd>Telescope keymaps<cr>", "[S]earch [K]eymaps" },
				l = { "<cmd>Telescope resume<cr>", "[S]earch [L]ast Query" },
				o = { "<cmd>Telescope live_grep grep_open_files=true<cr>", "[S]earch [O]pen Files" },
				p = { "<cmd>Telescope colorscheme enable_preview=true<cr>", "[S]earch Colorschemes With [P]review" },
				r = { "<cmd>Telescope oldfiles<cr>", "[S]earch [R]ecent Files" },
				s = { "<cmd>Telescope builtin<cr>", "[S]earch [S]elect Telescope" },
				t = { "<cmd>Telescope live_grep<cr>", "[S]earch [T]ext" },
				w = { "<cmd>Telescope grep_string<cr>", "[S]earch Current [W]ord" },
				["<leader>"] = { "<cmd>Telescope buffers<cr>", "[ ] Find Existing Buffers" }
			}
		},
		nopts = {
			buffer = nil,
			mode = "n",
			noremap = true,
			nowait = true,
			prefix = "<leader>",
			silent = true,
		},
		vmappings = {
			["/"] = { 
				"<Plug>(comment_toggle_linewise_visual)",
				"Comment toggle current line"
			},
		},
		vopts = {
			buffer = nil,
			mode = "v",
			noremap = true,
			nowait = true,
			prefix = "<leader>",
			silent = true,
		},
	}
end

M.setup = function()
	local wk = require("which-key")
		local c = config()
	
	wk.setup()
	wk.register(c.nmappings, c.nopts)
	wk.register(c.vmappings, c.vopts)
end


return M
