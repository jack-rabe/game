<script lang="ts">
	import { onMount } from 'svelte';

	const url = '://localhost:3333';
	let players: string[] = [];
	let user: string;
	let nameInput: HTMLInputElement;

	onMount(() => {
		fetch(`http${url}/getUsers`)
			.then((res) => res.json())
			.then((res) => (players = res));

		let message = '';
		let socket: WebSocket;
		const connectWebSocket = () => {
			socket = new WebSocket(`ws${url}/ws`);

			socket.onopen = (event) => {
				console.log('WebSocket connection opened', event);
				socket.send('hi');
			};
			socket.onmessage = (event) => {
				const newUser = JSON.parse(event.data);
				players.push(newUser.Name);
				players = players;
			};
			socket.onclose = (event) => {
				console.log('WebSocket connection closed', event);
			};
		};

		connectWebSocket();
	});
</script>

<h1 class="bg-primary font-bold text-2xl flex items-center justify-center h-14">Lobby</h1>
<div class="h-2/5 flex justify-center items-center">
	<div
		class="w-1/2 text-ellipsis overflow-hidden hover:overflow-visible h-40 bg-secondary flex rounded-lg"
	>
		{#each players as player}
			<span class="m-2 font-bold hover:animate-bounce">{player}</span>
		{/each}
	</div>
</div>
<div class="flex flex-col items-center m-2 h-2/5">
	<div class="m-4">
		<label for="name" class="label">Name</label>
		<input
			type="text"
			name="name"
			bind:value={user}
			bind:this={nameInput}
			class="input border-primary"
		/>
	</div>
	<button
		class="btn"
		on:click={() => {
			if (!user) {
				return;
			}
			const body = JSON.stringify({ Name: user });
			fetch(`http${url}/addUser`, {
				method: 'POST',
				body,
				headers: {
					'Content-Type': 'application/json'
				}
			});
			nameInput.value = '';
		}}>Play</button
	>
</div>

<style global>
	@import '../global.css';
</style>
