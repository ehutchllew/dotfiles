local M = {}

local function show_win_numbers()
	local wins = vim.api.nvim_list_wins()
	local floats = {}

	for i, win in ipairs(wins) do
		local float = vim.api.nvim_open_win(
			vim.api.nvim_create_buf(false, true), false, {
				relative = 'win',
				win = win,
				row = 0,
				col = 0,
				width = 3,
				height = 1,
				style = 'minimal',
				border = 'rounded',
				focusable = false,
			}
		)
		vim.api.nvim_buf_set_lines(
			vim.api.nvim_win_get_buf(float), 0, -1, false, { ' ' .. i .. ' ' }
		)
		table.insert(floats, float)
	end

	return floats
end

local function close_floats(floats)
	for _, float in ipairs(floats) do
		if vim.api.nvim_win_is_valid(float) then
			vim.api.nvim_win_close(float, true)
		end
	end
end

local function get_keypress(prompt)
	vim.api.nvim_echo({ { prompt, 'Normal' } }, false, {})
	vim.cmd('redraw')
	local ok, char = pcall(vim.fn.getcharstr)
	if not ok then return nil end
	local num = tonumber(char)
	return num
end

function M:init()
	-- Window Navigation
	vim.keymap.set("n", "<C-h>", "<C-w>h", { desc = "Move to Left Window" })
	vim.keymap.set("n", "<C-l>", "<C-w>l", { desc = "Move to Right Window" })
	vim.keymap.set("n", "<C-j>", "<C-w>j", { desc = "Move to Below Window" })
	vim.keymap.set("n", "<C-k>", "<C-w>k", { desc = "Move to Above Window" })

	-- Swapping Buffers Between Windows
	vim.keymap.set('n', '<space>wm', function()
		local wins = vim.api.nvim_list_wins()
		local floats = show_win_numbers()
		vim.cmd('redraw')

		local win1_num = get_keypress("First window: ")
		if not win1_num then
			close_floats(floats)
			return
		end
		local win1 = wins[win1_num]
		if not win1 then
			close_floats(floats)
			return
		end

		local win2_num = get_keypress("Second window: ")
		close_floats(floats)
		if not win2_num then return end
		local win2 = wins[win2_num]
		if not win2 or win2 == win1 then return end

		local buf1 = vim.api.nvim_win_get_buf(win1)
		local buf2 = vim.api.nvim_win_get_buf(win2)

		vim.api.nvim_win_set_buf(win1, buf2)
		vim.api.nvim_win_set_buf(win2, buf1)
	end, { desc = "Swap Buffers Between Two Prompted Windows" })
end

return M
