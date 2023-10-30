<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';

	const url = '://localhost:3333';
	let socket: WebSocket;
	let players = ['Player 1', 'Player 2', 'Player 3', 'Player 4'];
	let user: string;

	export let data: PageData;

	onMount(() => {
		const name = new URLSearchParams(window.location.search).get('name');
		user = name || '';

		socket = new WebSocket(`ws${url}/joinGame/${data.gameId}?name=${name}`);
		socket.onopen = (event) => {
			console.log('WebSocket connection opened', event);
			socket.send('hi');
		};
		socket.onmessage = (event) => {
			const data = JSON.parse(event.data);
			if (data.type === 'load') {
				players = data.players;
			}
			console.log(data);
		};
		socket.onclose = (event) => {
			alert('Lost connection to server!');
			console.log('WebSocket connection closed', event);
		};
	});
</script>

<div class=" fixed top-0 m-2 flex w-full justify-between text-xl">
	<div class="rounded-sm bg-neutral px-3 py-2">{players[0]}</div>
	<div class="mr-4 rounded-sm bg-neutral px-3 py-2">{players[1]}</div>
</div>
<div class="h-20" />
<div class="mx-3 h-3/4 rounded-lg bg-neutral" />
<div class="fixed bottom-0 m-2 flex w-full justify-between text-xl">
	<div class="rounded-sm bg-neutral px-3 py-2">{players[2]}</div>
	<!-- TODO why do i need to add this weird margin below? -->
	<div class="mr-4 rounded-sm bg-neutral px-3 py-2">{players[3]}</div>
</div>
