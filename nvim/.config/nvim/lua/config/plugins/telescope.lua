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
		require("telescope").setup({})
		vim.keymap.set("n", "<space>sf", require("telescope.builtin").find_files)
		vim.keymap.set("n", "<space>st", require("telescope.builtin").live_grep)
		vim.keymap.set("n", "<space>sl", require("telescope.builtin").resume)
	end
}
