local M = {}

-- Two important keymaps to use while in Telescope are:
-- - Insert mode: <c-/>
-- - Normal mode: ?
--
-- This opens a window that show you all keymaps for the
-- current telescope picker. Useful to discover what 
-- Telescope can do as well as how to do it.

local function config()
	return {
		extensions = {
			["ui-select"] = {
				require("telescope.themes").get_dropdown()
			}
		}
	}
end

M.setup = function()
	local ts = require("telescope")
	local c = config()
	
	ts.setup(c)

	-- Enable extensions if installed
	pcall(ts.load_extension, "fzf")
	pcall(ts.load_extension, "ui-select")

	-- See: `:help telescope.builtin`
	-- local tbi = require("telescope.builtin")
	-- vim.keymap.set('n', '<leader>sh', builtin.help_tags, { desc = '[S]earch [H]elp' })
 --      	vim.keymap.set('n', '<leader>sk', builtin.keymaps, { desc = '[S]earch [K]eymaps' })
 --      	vim.keymap.set('n', '<leader>sf', builtin.find_files, { desc = '[S]earch [F]iles' })
 --      	vim.keymap.set('n', '<leader>ss', builtin.builtin, { desc = '[S]earch [S]elect Telescope' })
 --      	vim.keymap.set('n', '<leader>sw', builtin.grep_string, { desc = '[S]earch current [W]ord' })
 --      	vim.keymap.set('n', '<leader>st', builtin.live_grep, { desc = '[S]earch by [T]ext' })
 --      	vim.keymap.set('n', '<leader>sd', builtin.diagnostics, { desc = '[S]earch [D]iagnostics' })
 --      	vim.keymap.set('n', '<leader>sl', builtin.resume, { desc = '[S]earch [L]ast Query' })
 --      	vim.keymap.set('n', '<leader>sr', builtin.oldfiles, { desc = '[S]earch [R]ecent Files ("." for repeat)' })
 --      	vim.keymap.set('n', '<leader><leader>', builtin.buffers, { desc = '[ ] Find existing buffers' })
end


return M

