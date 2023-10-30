<script lang="ts">
	import { afterUpdate, onMount } from 'svelte';
	import type { PageData } from './$types';

	type Message = {
		sender: string;
		content: string;
	};

	const url = '://localhost:3333';
	let socket: WebSocket;
	let players = ['Player 1', 'Player 2', 'Player 3', 'Player 4'];
	let user: string;
	let messages: Message[] = [];
	let messagesElement: HTMLElement;

	export let data: PageData;

	onMount(() => {
		const name = new URLSearchParams(window.location.search).get('name');
		user = name || '';

		socket = new WebSocket(`ws${url}/joinGame/${data.gameId}?name=${name}`);
		socket.onopen = (event) => {
			console.log('WebSocket connection opened', event);
			const msg: Message = { sender: user, content: `${user} joined the game` };
			socket.send(JSON.stringify(msg));
		};
		socket.onmessage = (event) => {
			const data = JSON.parse(event.data);
			if (data.type === 'load') {
				players = data.players;
			} else if (data.type === 'message') {
				messages.push(data.message);
				messages = messages;
			}
			console.log(data);
		};
		socket.onclose = (event) => {
			alert('Lost connection to server!');
			console.log('WebSocket connection closed', event);
		};
	});

	// scroll to the bottom so that the newest messages are seen
	afterUpdate(() => {
		if (messages) scrollToBottom(messagesElement);
	});
	const scrollToBottom = async (node: HTMLElement) => {
		node.scroll({ top: node.scrollHeight, behavior: 'smooth' });
	};
</script>

<div class="fixed top-0 m-2 flex w-full justify-between text-xl">
	<div class="rounded-sm bg-neutral px-3 py-2">{players[0]}</div>
	<div class="mr-4 rounded-sm bg-neutral px-3 py-2">{players[1]}</div>
</div>
<div class="h-20" />
<div class="mx-3 grid h-3/4 grid-cols-5">
	<div class="col-span-4 mr-2 rounded-lg bg-neutral">main game</div>
	<div class="rounded-lg bg-neutral p-1">
		<div bind:this={messagesElement} class="h-[34rem] overflow-y-scroll">
			{#each messages as msg}
				{#if user === msg.sender}
					<div class="chat chat-start">
						<div class="chat-header">{msg.sender}</div>
						<div class="chat-bubble chat-bubble-primary">{msg.content}</div>
					</div>
				{:else}
					<div class="chat chat-end">
						<div class="chat-header">{msg.sender}</div>
						<div class="chat-bubble chat-bubble-secondary">{msg.content}</div>
					</div>
				{/if}
			{/each}
		</div>
		<div class="flex h-1/6 items-center justify-center rounded-lg">
			<!-- TODO scroll to bottom -->
			<input type="text" class="input mr-1" />
			<button class="btn">Send</button>
		</div>
	</div>
</div>
<div class="fixed bottom-0 m-2 flex w-full justify-between text-xl">
	<div class="rounded-sm bg-neutral px-3 py-2">{players[2]}</div>
	<!-- TODO why do i need to add this weird margin below? -->
	<div class="mr-4 rounded-sm bg-neutral px-3 py-2">{players[3]}</div>
</div>
