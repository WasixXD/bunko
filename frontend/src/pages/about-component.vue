
<template>
    Hello, About


    vue: {{ hello() }}


    <Button class="cursor-pointer">Hello guys</Button>
    
</template>

<script setup lang="ts">
    import { ref } from 'vue';
    import { Button } from '@/components/ui/button';

    
    console.log("Hello World")



    function hello(): string {
        return "Hello"
    }

</script>