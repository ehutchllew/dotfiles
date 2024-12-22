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
			extensions = {
				fzf = {}
			}
		})

		require("telescope").load_extension("fzf")
		vim.keymap.set("n", "<space>sf", require("telescope.builtin").find_files)
		vim.keymap.set("n", "<space>st", require("telescope.builtin").live_grep)
		vim.keymap.set("n", "<space>sl", require("telescope.builtin").resume)

		-- Keymap to search all plugin files... for science
		vim.keymap.set("n", "<space>sp", function()
			require("telescope.builtin").find_files({
				cwd = vim.fs.joinpath(vim.fn.stdpath("data"), "lazy")
			})
		end)

		require("config.telescope.multigrep").setup()
	end
}
