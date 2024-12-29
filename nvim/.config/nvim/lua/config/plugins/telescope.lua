return {
	'nvim-telescope/telescope.nvim',
	tag = '0.1.8',
	dependencies = {
		'nvim-lua/plenary.nvim',
		{
			'nvim-telescope/telescope-fzf-native.nvim',
			build =
			'cmake -S. -Bbuild -DCMAKE_BUILD_TYPE=Release && cmake --build build --config Release'
		}
	},
	config = function()
		require("telescope").setup({
			defaults = {
				vimgrep_arguments = {
					"rg",
					"--color=never",
					"--no-heading",
					"--with-filename",
					"--line-number",
					"--column",
					"--smart-case",
					"--hidden",
					"--glob=!.git/",
				},
			},
			extensions = {
				fzf = {}
			},
			pickers = {
				find_files = {
					find_command = { "rg", "--files", "--hidden", "--glob", "!**/.git/*" }
				}
			}
		})

		require("telescope").load_extension("fzf")
		vim.keymap.set("n", "<space>sc", require("telescope.builtin").commands, { desc = "Search Available Commands" })
		vim.keymap.set("n", "<space>sd", require("telescope.builtin").diagnostics, { desc = "Search Diagnostics" })
		vim.keymap.set("n", "<space>sf", require("telescope.builtin").find_files, { desc = "Search Files" })
		vim.keymap.set("n", "<space>sk", require("telescope.builtin").keymaps, { desc = "Search Keymaps" })
		vim.keymap.set("n", "<space>st", require("telescope.builtin").live_grep, { desc = "Search Text" })
		vim.keymap.set("n", "<space>sl", require("telescope.builtin").resume, { desc = "Search Last Query" })

		-- Keymap to search all plugin files... for science
		vim.keymap.set("n", "<space>sp", function()
				require("telescope.builtin").find_files({
					cwd = vim.fs.joinpath(vim.fn.stdpath("data"), "lazy")
				})
			end,
			{ desc = "Search Plugin Files" })

		require("config.telescope.multigrep").setup()
	end
}
